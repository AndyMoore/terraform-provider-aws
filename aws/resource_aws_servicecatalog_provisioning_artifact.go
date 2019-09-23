package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogProvisioningArtifact() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProvisioningArtifactCreate,
		Read:   resourceAwsServiceCatalogProvisioningArtifactRead,
		Update: resourceAwsServiceCatalogProvisioningArtifactUpdate,
		Delete: resourceAwsServiceCatalogProvisioningArtifactDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate,
			}, // CLOUD_FORMATION_TEMPLATE  | MARKETPLACE_AMI | MARKETPLACE_CA
			"template_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsServiceCatalogProvisioningArtifactCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := servicecatalog.CreateProvisioningArtifactInput{
		IdempotencyToken: aws.String(resource.UniqueId()),
	}

	productId := d.Get("product_id").(string)
	input.ProductId = aws.String(productId)

	artifactProperties := servicecatalog.ProvisioningArtifactProperties{}
	artifactProperties.Description = aws.String(d.Get("description").(string))
	artifactProperties.Name = aws.String(d.Get("name").(string))

	if v, ok := d.GetOk("type"); ok && v != "" {
		artifactProperties.Type = aws.String(v.(string))
	} else {
		artifactProperties.Type = aws.String(servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate)
	}
	artifactProperties.Info = make(map[string]*string)
	artifactProperties.Info["LoadTemplateFromURL"] = aws.String(d.Get("template_url").(string))

	input.SetParameters(&artifactProperties)
	log.Printf("[DEBUG] Creating Service Catalog Product Artifact: %s %s", input, artifactProperties)

	resp, err := conn.CreateProvisioningArtifact(&input)
	if resp != nil {
		fmt.Errorf("Unable to provision artifact: %s", err)
	}

	artifactId := aws.StringValue(resp.ProvisioningArtifactDetail.Id)

	waitForCreated := &resource.StateChangeConf{
		Target: []string{servicecatalog.StatusAvailable},
		Refresh: func() (result interface{}, state string, err error) {
			resp, err := conn.DescribeProvisioningArtifact(&servicecatalog.DescribeProvisioningArtifactInput{
				ProductId:              aws.String(productId),
				ProvisioningArtifactId: aws.String(artifactId),
			})
			if err != nil {
				return nil, "", err
			}
			return resp, aws.StringValue(resp.Status), nil
		},
		Timeout:      d.Timeout(schema.TimeoutCreate),
		PollInterval: 3 * time.Second,
	}
	if _, err := waitForCreated.WaitForState(); err != nil {
		return err
	}

	d.SetId(artifactId)

	return resourceAwsServiceCatalogProvisioningArtifactRead(d, meta)
}

func resourceAwsServiceCatalogProvisioningArtifactRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	resp, err := conn.DescribeProvisioningArtifact(&servicecatalog.DescribeProvisioningArtifactInput{
		ProductId:              aws.String(d.Get("product_id").(string)),
		ProvisioningArtifactId: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Service Catalog Provisioned Product %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading ServiceCatalog provisioned product '%s' failed: %s", d.Id(), err)
	}

	pad := resp.ProvisioningArtifactDetail

	d.Set("description", pad.Description)
	d.Set("name", pad.Name)
	d.Set("type", pad.Type)
	d.Set("template_url", aws.StringValue(resp.Info["TemplateUrl"]))
	d.Set("created_time", pad.CreatedTime.Format(time.RFC3339))
	d.Set("active", pad.Active)
	d.Set("status", resp.Status)

	return nil
}

func resourceAwsServiceCatalogProvisioningArtifactUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := servicecatalog.UpdateProvisioningArtifactInput{
		ProvisioningArtifactId: aws.String(d.Id()),
		ProductId:              aws.String(d.Get("product_id").(string)),
	}

	if d.HasChange("description") {
		v, _ := d.GetOk("name")
		input.Description = aws.String(v.(string))
	}

	if d.HasChange("active") {
		v, _ := d.GetOk("active")
		input.Active = aws.Bool(v.(bool))
	}

	_, err := conn.UpdateProvisioningArtifact(&input)

	if err != nil {
		fmt.Errorf("Error updating artifact %s: %s", aws.String(d.Id()), err)
	}

	return resourceAwsServiceCatalogProvisioningArtifactRead(d, meta)
}

func resourceAwsServiceCatalogProvisioningArtifactDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DeleteProvisioningArtifactInput{}
	input.ProvisioningArtifactId = aws.String(d.Id())
	input.ProductId = aws.String(d.Get("product_id").(string))

	log.Printf("[DEBUG] Delete Service Catalog Provisioning Artifact: %s", input)
	_, err := conn.DeleteProvisioningArtifact(&input)
	if err != nil {
		return fmt.Errorf("deleting ServiceCatalog Provisioning Artifact '%s' failed: %s", *input.ProvisioningArtifactId, err)
	}
	return nil
}
