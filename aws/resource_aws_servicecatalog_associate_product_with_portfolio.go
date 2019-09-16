package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/hashcode"
	//"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogAssociateProductWithPortfolio() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogAssociateProductWithPortfolioCreate,
		Read:   resourceAwsServiceCatalogAssociateProductWithPortfolioRead,
		Update: resourceAwsServiceCatalogAssociateProductWithPortfolioUpdate,
		Delete: resourceAwsServiceCatalogAssociateProductWithPortfolioDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source_portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}
func resourceAwsServiceCatalogAssociateProductWithPortfolioCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.AssociateProductWithPortfolioInput{
		AcceptLanguage:    aws.String("en"),
		PortfolioId:       aws.String(d.Get("portfolio_id").(string)),
		ProductId:         aws.String(d.Get("product_id").(string)),
		SourcePortfolioId: aws.String(d.Get("source_portfolio_id").(string)),
	}

	log.Printf("[DEBUG] Creating Service Catalog Portfolio: %#v", input)
	resp, err := conn.AssociateProductWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Creating Associating Service Catalog Product with Portfolio failed: %s", err.Error())
	}

	id := productPortfolioIDHash(d)

	d.SetId(id)

	return nil
}

func resourceAwsServiceCatalogAssociateProductWithPortfolioRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	portfolio_id = d.Get("portfolio_id").(string)
	product_id = d.Get("product_id").(string)

	input := servicecatalog.ListPortfoliosForProductInput{
		AcceptLanguage: aws.String("en"),
		ProductId:      product_id,
	}

	resp, err := svc.ListPortfoliosForProduct(input)
	if err != nil {
		fmt.Errorf("error listing portfolios for product: %s", err.Error())
	}

	var found = 0

	for _, p := range resp.PortfolioDetails {
		if aws.StringValue(p.Id) == portfolio_id {
			found := 1
			//fmt.Printf(aws.StringValue(p.Id))
			//fmt.Println("p ", reflect.TypeOf(p))
		}
	}

	if found == 0 {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsServiceCatalogAssociateProductWithPortfolioUpdate(d *schema.ResourceData, meta interface{}) error {
        conn := meta.(*AWSClient).scconn
        product_id = d.Get("product_id").(string)

        input := servicecatalog.AssociateProductWithPortfolioInput{
                AcceptLanguage:    aws.String("en"),
                PortfolioId:       aws.String(d.Get("portfolio_id").(string)),
                ProductId:         aws.String(d.Get("product_id").(string)),
                SourcePortfolioId: aws.String(d.Get("source_portfolio_id").(string)),
        }

        if d.HasChange("accept_language") {
                v, _ := d.GetOk("accept_language")
                input.AcceptLanguage = aws.String(v.(string))
        }

        if d.HasChange("product_id") {
                v, _ := d.GetOk("provider_name")
                input.ProductId = aws.String(v.(string))
        }

	DisassociateProductFromPortfolio


}

func resourceAwsServiceCatalogAssociateProductWithPortfolioDelete(d *schema.ResourceData, meta interface{}) error {
}


func productPortfolioIDHash(d *schema.ResourceData) string {
	var buf bytes.Buffer
	portfolio_id := d.Get("portfolio_id").(string)
	product_id := d.Get("portfolio_id").(string)

	buf.WriteString(fmt.Sprintf("%s-%s", portfolio_id, product_id))

	return fmt.Sprintf("product-portfolio-%d", hashcode.String(buf.String()))
}
