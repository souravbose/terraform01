package streamanalytics

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-sdk/resource-manager/streamanalytics/2020-03-01/inputs"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/streamanalytics/migration"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceStreamAnalyticsStreamInputEventHub() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceStreamAnalyticsStreamInputEventHubCreateUpdate,
		Read:   resourceStreamAnalyticsStreamInputEventHubRead,
		Update: resourceStreamAnalyticsStreamInputEventHubCreateUpdate,
		Delete: resourceStreamAnalyticsStreamInputEventHubDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := inputs.ParseInputID(id)
			return err
		}),

		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.StreamAnalyticsStreamInputEventHubV0ToV1{},
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"stream_analytics_job_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"eventhub_consumer_group_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"eventhub_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"servicebus_namespace": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"shared_access_policy_key": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"shared_access_policy_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"partition_key": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"authentication_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(inputs.AuthenticationModeConnectionString),
				ValidateFunc: validation.StringInSlice([]string{
					string(inputs.AuthenticationModeMsi),
					string(inputs.AuthenticationModeConnectionString),
				}, false),
			},

			"serialization": schemaStreamAnalyticsStreamInputSerialization(),
		},
	}
}

func resourceStreamAnalyticsStreamInputEventHubCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.InputsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure Stream Analytics Stream Input EventHub creation.")
	id := inputs.NewInputID(subscriptionId, d.Get("resource_group_name").(string), d.Get("stream_analytics_job_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id)
		if err != nil {
			if !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !response.WasNotFound(existing.HttpResponse) {
			return tf.ImportAsExistsError("azurerm_stream_analytics_stream_input_eventhub", id.ID())
		}
	}

	serializationRaw := d.Get("serialization").([]interface{})
	serialization, err := expandStreamAnalyticsStreamInputSerialization(serializationRaw)
	if err != nil {
		return fmt.Errorf("expanding `serialization`: %+v", err)
	}

	eventHubDataSourceProps := &inputs.EventHubStreamInputDataSourceProperties{
		EventHubName:        utils.String(d.Get("eventhub_name").(string)),
		ServiceBusNamespace: utils.String(d.Get("servicebus_namespace").(string)),
		ConsumerGroupName:   utils.String(d.Get("eventhub_consumer_group_name").(string)),
		AuthenticationMode:  utils.ToPtr(inputs.AuthenticationMode(d.Get("authentication_mode").(string))),
	}

	if v, ok := d.GetOk("shared_access_policy_key"); ok {
		eventHubDataSourceProps.SharedAccessPolicyKey = utils.String(v.(string))
	}

	if v, ok := d.GetOk("shared_access_policy_name"); ok {
		eventHubDataSourceProps.SharedAccessPolicyName = utils.String(v.(string))
	}

	var dataSource inputs.StreamInputDataSource = inputs.EventHubStreamInputDataSource{
		Properties: eventHubDataSourceProps,
	}
	var inputProperties inputs.InputProperties = inputs.StreamInputProperties{
		Datasource:    pointer.To(dataSource),
		Serialization: pointer.To(serialization),
		PartitionKey:  utils.String(d.Get("partition_key").(string)),
	}
	props := inputs.Input{
		Name:       utils.String(id.InputName),
		Properties: &inputProperties,
	}

	if d.IsNewResource() {
		if _, err := client.CreateOrReplace(ctx, id, props, inputs.DefaultCreateOrReplaceOperationOptions()); err != nil {
			return fmt.Errorf("creating %s: %+v", id, err)
		}

		d.SetId(id.ID())
	} else if _, err := client.Update(ctx, id, props, inputs.DefaultUpdateOperationOptions()); err != nil {
		return fmt.Errorf("updating %s: %+v", id, err)
	}

	return resourceStreamAnalyticsStreamInputEventHubRead(d, meta)
}

func resourceStreamAnalyticsStreamInputEventHubRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.InputsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := inputs.ParseInputID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, *id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			log.Printf("[DEBUG] %s was not found - removing from state!", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.InputName)
	d.Set("stream_analytics_job_name", id.StreamingJobName)
	d.Set("resource_group_name", id.ResourceGroupName)

	if model := resp.Model; model != nil {
		if props := model.Properties; props != nil {
			if input, ok := (*props).(inputs.InputProperties); ok {
				if streamInput, ok := input.(inputs.StreamInputProperties); ok {
					if ds := streamInput.Datasource; ds != nil {
						if streamEventHubInput, ok := (*ds).(inputs.EventHubStreamInputDataSource); ok {
							if inputProps := streamEventHubInput.Properties; inputProps != nil {
								eventHubName := ""
								if v := inputProps.EventHubName; v != nil {
									eventHubName = *v
								}
								d.Set("eventhub_name", eventHubName)

								serviceBusNameSpace := ""
								if v := inputProps.ServiceBusNamespace; v != nil {
									serviceBusNameSpace = *v
								}
								d.Set("servicebus_namespace", serviceBusNameSpace)

								authMode := ""
								if v := inputProps.AuthenticationMode; v != nil {
									authMode = string(*v)
								}
								d.Set("authentication_mode", authMode)

								consumerGroupName := ""
								if v := inputProps.ConsumerGroupName; v != nil {
									consumerGroupName = *v
								}
								d.Set("eventhub_consumer_group_name", consumerGroupName)

								sharedAccessPolicyName := ""
								if v := inputProps.SharedAccessPolicyName; v != nil {
									sharedAccessPolicyName = *v
								}
								d.Set("shared_access_policy_name", sharedAccessPolicyName)

								partitionKey := ""
								if v := streamInput.PartitionKey; v != nil {
									partitionKey = *v
								}
								d.Set("partition_key", partitionKey)
							}
						}
					}

					if err := d.Set("serialization", flattenStreamAnalyticsStreamInputSerialization(streamInput.Serialization)); err != nil {
						return fmt.Errorf("setting `serialization`: %+v", err)
					}
				}
			}
		}
	}

	return nil
}

func resourceStreamAnalyticsStreamInputEventHubDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.InputsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := inputs.ParseInputID(d.Id())
	if err != nil {
		return err
	}

	if resp, err := client.Delete(ctx, *id); err != nil {
		if !response.WasNotFound(resp.HttpResponse) {
			return fmt.Errorf("deleting %s: %+v", *id, err)
		}
	}

	return nil
}
