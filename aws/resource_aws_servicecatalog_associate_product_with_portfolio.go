package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/customdiff"
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
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				importParts, err := validateAssociateImportString(d.Id())
				if err != nil {
					return nil, err
				}
				if err := populateAssociationFromImport(d, importParts); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			// This is to prevent this error:
			// All fields are ForceNew or Computed w/out Optional, Update is superfluous
			customdiff.ForceNewIfChange("product_id", func(old, new, meta interface{}) bool {
				return old.(string) != new.(string)
			}),
			customdiff.ForceNewIfChange("source_portfolio_id", func(old, new, meta interface{}) bool {
				return old.(string) != new.(string)
			}),
		),

		Schema: map[string]*schema.Schema{
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source_portfolio_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsServiceCatalogAssociateProductWithPortfolioCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.AssociateProductWithPortfolioInput{
		AcceptLanguage: aws.String("en"),
		PortfolioId:    aws.String(d.Get("portfolio_id").(string)),
		ProductId:      aws.String(d.Get("product_id").(string)),
	}

	if v, ok := d.GetOk("source_portfolio_id"); ok {
		input.SourcePortfolioId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Service Catalog Portfolio: %#v", input)
	_, err := conn.AssociateProductWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Associating Service Catalog Product with Portfolio failed: %s", err.Error())
	}

	return resourceAwsServiceCatalogAssociateProductWithPortfolioRead(d, meta)
}

func resourceAwsServiceCatalogAssociateProductWithPortfolioRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	portfolio_id := d.Get("portfolio_id").(string)
	product_id := d.Get("product_id").(string)

	input := servicecatalog.ListPortfoliosForProductInput{
		AcceptLanguage: aws.String("en"),
		ProductId:      aws.String(product_id),
	}

	resp, err := conn.ListPortfoliosForProduct(&input)
	if err != nil {
		fmt.Errorf("error listing portfolios for product: %s", err.Error())
	}

	var found = 0

	for _, p := range resp.PortfolioDetails {
		if aws.StringValue(p.Id) == portfolio_id {
			found = 1
		}
	}

	if found == 0 {
		d.SetId("")
		return nil
	} else {
		id := productPortfolioIDHash(d)
		d.SetId(id)
	}
	return nil
}

func resourceAwsServiceCatalogAssociateProductWithPortfolioUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsServiceCatalogAssociateProductWithPortfolioRead(d, meta)
}

func validateAssociateImportString(importStr string) ([]string, error) {
	// example: port-ig54mbjew7qru_prod-z2koxglqdw4n4

	log.Printf("[DEBUG] Validating import string %s", importStr)

	importParts := strings.Split(strings.ToLower(importStr), "_")
	errStr := "unexpected format of import string (%q), expected PORTFOLIOID_PRODUCTID: %s"
	if len(importParts) != 2 {
		return nil, fmt.Errorf(errStr, importStr, "too few parts")
	}

	portfolioId := importParts[0]
	productId := importParts[1]

	if !strings.HasPrefix(portfolioId, "port-") {
		return nil, fmt.Errorf(errStr, importStr, "invalid portfolio ID")
	}

	if !strings.HasPrefix(productId, "prod-") {
		return nil, fmt.Errorf(errStr, importStr, "invalid product ID")
	}

	log.Printf("[DEBUG] Validated import string %s", importStr)
	return importParts, nil
}

func populateAssociationFromImport(d *schema.ResourceData, importParts []string) error {
	log.Printf("[DEBUG] Populating resource data on import: %v", importParts)

	portfolioId := importParts[0]
	productId := importParts[1]

	d.Set("portfolio_id", portfolioId)
	d.Set("product_id", productId)

	d.SetId(productPortfolioIDHash(d))

	return nil
}

func resourceAwsServiceCatalogAssociateProductWithPortfolioDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := servicecatalog.DisassociateProductFromPortfolioInput{
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
		ProductId:   aws.String(d.Get("product_id").(string)),
	}

	log.Printf("[DEBUG] Disassociate product from portfolio: %#v", input)

	_, err := conn.DisassociateProductFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Disassociating Product (%s) from Portfolio (%s) failed: %s", *input.ProductId, *input.PortfolioId, err.Error())
	}
	return nil
}

func productPortfolioIDHash(d *schema.ResourceData) string {
	var buf bytes.Buffer
	portfolio_id := d.Get("portfolio_id").(string)
	product_id := d.Get("portfolio_id").(string)

	buf.WriteString(fmt.Sprintf("%s-%s", portfolio_id, product_id))

	return fmt.Sprintf("product-portfolio-%d", hashcode.String(buf.String()))
}
