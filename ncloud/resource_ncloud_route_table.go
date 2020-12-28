package ncloud

import (
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func init() {
	RegisterResource("ncloud_route_table", resourceNcloudRouteTable())
}

func resourceNcloudRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceNcloudRouteTableCreate,
		Read:   resourceNcloudRouteTableRead,
		Update: resourceNcloudRouteTableUpdate,
		Delete: resourceNcloudRouteTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: ncloudVpcCommonCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"vpc_no": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"supported_subnet_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ToDiagFunc(validation.StringInSlice([]string{"PUBLIC", "PRIVATE"}, false)),
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: ToDiagFunc(validateInstanceName),
			},
			"description": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: ToDiagFunc(validation.StringLenBetween(0, 1000)),
			},
			"route_table_no": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceNcloudRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	if !config.SupportVPC {
		return NotSupportClassic("resource `ncloud_route_table`")
	}

	reqParams := &vpc.CreateRouteTableRequest{
		RegionCode:              &config.RegionCode,
		VpcNo:                   ncloud.String(d.Get("vpc_no").(string)),
		SupportedSubnetTypeCode: ncloud.String(d.Get("supported_subnet_type").(string)),
	}

	if v, ok := d.GetOk("name"); ok {
		reqParams.RouteTableName = ncloud.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		reqParams.RouteTableDescription = ncloud.String(v.(string))
	}

	logCommonRequest("CreateRouteTable", reqParams)
	resp, err := config.Client.vpc.V2Api.CreateRouteTable(reqParams)
	if err != nil {
		logErrorResponse("CreateRouteTable", err, reqParams)
		return err
	}

	logResponse("CreateRouteTable", resp)

	instance := resp.RouteTableList[0]
	d.SetId(*instance.RouteTableNo)

	log.Printf("[INFO] Route Table ID: %s", d.Id())

	if err := waitForNcloudRouteTableCreation(config, d.Id()); err != nil {
		return err
	}

	return resourceNcloudRouteTableRead(d, meta)
}

func resourceNcloudRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	instance, err := getRouteTableInstance(config, d.Id())
	if err != nil {
		return err
	}

	if instance == nil {
		d.SetId("")
		return nil
	}

	d.SetId(*instance.RouteTableNo)
	d.Set("route_table_no", instance.RouteTableNo)
	d.Set("name", instance.RouteTableName)
	d.Set("description", instance.RouteTableDescription)
	d.Set("vpc_no", instance.VpcNo)
	d.Set("supported_subnet_type", instance.SupportedSubnetType.Code)
	d.Set("is_default", instance.IsDefault)

	return nil
}

func resourceNcloudRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	if d.HasChange("description") {
		if err := setRouteTableDescription(d, config); err != nil {
			return err
		}
	}

	return resourceNcloudRouteTableRead(d, meta)
}

func resourceNcloudRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	reqParams := &vpc.DeleteRouteTableRequest{
		RegionCode:   &config.RegionCode,
		RouteTableNo: ncloud.String(d.Get("route_table_no").(string)),
	}

	logCommonRequest("DeleteRouteTable", reqParams)
	resp, err := config.Client.vpc.V2Api.DeleteRouteTable(reqParams)
	if err != nil {
		logErrorResponse("DeleteRouteTable", err, reqParams)
		return err
	}

	logResponse("DeleteRouteTable", resp)

	if err := waitForNcloudRouteTableDeletion(config, d.Id()); err != nil {
		return err
	}

	return nil
}

func waitForNcloudRouteTableCreation(config *ProviderConfig, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"INIT", "CREATING"},
		Target:  []string{"RUN"},
		Refresh: func() (interface{}, string, error) {
			instance, err := getRouteTableInstance(config, id)
			return VpcCommonStateRefreshFunc(instance, err, "RouteTableStatus")
		},
		Timeout:    DefaultTimeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for Route Table (%s) to become running: %s", id, err)
	}

	return nil
}

func waitForNcloudRouteTableDeletion(config *ProviderConfig, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"RUN", "TERMTING"},
		Target:  []string{"TERMINATED"},
		Refresh: func() (interface{}, string, error) {
			instance, err := getRouteTableInstance(config, id)
			return VpcCommonStateRefreshFunc(instance, err, "RouteTableStatus")
		},
		Timeout:    DefaultTimeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for Route Table (%s) to become termintaing: %s", id, err)
	}

	return nil
}

func getRouteTableInstance(config *ProviderConfig, id string) (*vpc.RouteTable, error) {
	reqParams := &vpc.GetRouteTableDetailRequest{
		RegionCode:   &config.RegionCode,
		RouteTableNo: ncloud.String(id),
	}

	logCommonRequest("GetRouteTableDetail", reqParams)
	resp, err := config.Client.vpc.V2Api.GetRouteTableDetail(reqParams)
	if err != nil {
		logErrorResponse("GetRouteTableDetail", err, reqParams)
		return nil, err
	}
	logResponse("GetRouteTableDetail", resp)

	if len(resp.RouteTableList) > 0 {
		return resp.RouteTableList[0], nil
	}

	return nil, nil
}

func setRouteTableDescription(d *schema.ResourceData, config *ProviderConfig) error {
	reqParams := &vpc.SetRouteTableDescriptionRequest{
		RegionCode:            &config.RegionCode,
		RouteTableNo:          ncloud.String(d.Id()),
		RouteTableDescription: StringPtrOrNil(d.GetOk("description")),
	}

	logCommonRequest("setRouteTableDescription", reqParams)
	resp, err := config.Client.vpc.V2Api.SetRouteTableDescription(reqParams)
	if err != nil {
		logErrorResponse("setRouteTableDescription", err, reqParams)
		return err
	}
	logResponse("setRouteTableDescription", resp)

	return nil
}
