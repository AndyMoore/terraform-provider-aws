package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsServiceCatalogProduct() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceCatalogProductRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"distributor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"has_default_path": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"product_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsServiceCatalogProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	productId := d.Get("id").(string)

	input := servicecatalog.DescribeProductAsAdminInput{
		Id: aws.String(productId),
	}

	output, err := conn.DescribeProductAsAdmin(&input)
	if err != nil {
		return fmt.Errorf("Unable to find product %s: %s", productId, err)
	}

	pvd := output.ProductViewDetail
	pvs := pvd.ProductViewSummary

	if pvs.ShortDescription != nil {
		d.Set("description", pvs.ShortDescription)
	}
	if pvs.ShortDescription != nil {
		d.Set("description", pvs.ShortDescription)
	}
	if pvs.Distributor != nil {
		d.Set("distributor", pvs.Distributor)
	}
	if pvs.Name != nil {
		d.Set("name", pvs.Name)
	}
	if pvs.Owner != nil {
		d.Set("owner", pvs.Owner)
	}
	if pvd.ProductARN != nil {
		d.Set("product_arn", pvd.ProductARN)
	}
	if pvs.HasDefaultPath != nil {
		d.Set("has_default_path", pvs.HasDefaultPath)
	}
	if pvs.Type != nil {
		d.Set("product_type", pvs.Type)
	}
	if pvs.SupportDescription != nil {
		d.Set("support_description", pvs.SupportDescription)
	}
	if pvs.SupportEmail != nil {
		d.Set("support_email", pvs.SupportEmail)
	}
	if pvs.SupportUrl != nil {
		d.Set("support_url", pvs.SupportUrl)
	}

	d.Set("tags", tagsToMapServiceCatalog(output.Tags))

	d.SetId(aws.StringValue(pvs.ProductId))

	return nil
}
