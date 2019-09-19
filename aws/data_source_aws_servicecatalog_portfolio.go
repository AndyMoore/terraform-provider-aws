package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsServiceCatalogPortfolio() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceCatalogPortfolioRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"provider_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsServiceCatalogPortfolioRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	portfolio_id := d.Get("id").(string)

	input := servicecatalog.DescribePortfolioInput{
		Id: aws.String(portfolio_id),
	}

	output, err := conn.DescribePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Unable to find portfolio %s: %s", portfolio_id, err)
	}

	pd := output.PortfolioDetail

	d.Set("arn", pd.ARN)
	d.Set("create_time", pd.CreatedTime)
	d.Set("description", pd.Description)
	d.Set("name", pd.DisplayName)
	d.Set("provider_name", pd.ProviderName)
	d.Set("tags", tagsToMapSC(output.Tags))

	d.SetId(pd.Id)

	return nil
}
