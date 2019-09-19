package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSServiceCatalogAssociateProductWithPortfolio_basic(t *testing.T) {
	productName := "product1"
	portfolioName := "portfolio1"
	portfolioName2 := "portfolio2"
	provisionArtifactName := "artifact1"
	bucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	resourceName := "aws_servicecatalog_associate_product_with_portfolio.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogAssociateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogResourceConfigTemplate(bucketName, portfolioName, portfolioName2, productName, provisionArtifactName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogAssociateExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogAssociateProductWithPortfolio_update(t *testing.T) {
	productName := "product1"
	portfolioName := "portfolio1"
	portfolioName2 := "portfolio2"
	provisionArtifactName := "artifact1"
	bucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	resourceName := "aws_servicecatalog_associate_product_with_portfolio.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogAssociateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogResourceConfigTemplate(bucketName, portfolioName, portfolioName2, productName, provisionArtifactName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogAssociateExists(resourceName),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogResourceConfigUpdateTemplate(bucketName, portfolioName, portfolioName2, productName, provisionArtifactName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogAssociateExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSAssociateProductWithPortfolioImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

//func TestAccAWSServiceCatalogAssociateProductWithPortfolio_import(t *testing.T) {

func testAccCheckServiceCatalogAssociateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_associate_product_with_portfolio" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn
		productId := rs.Primary.Attributes["product_id"]
		resp, err := conn.ListPortfoliosForProduct(&servicecatalog.ListPortfoliosForProductInput{
			ProductId: aws.String(productId),
		})

		if err == nil {
			if len(resp.PortfolioDetails) != 0 {
				return fmt.Errorf("Association %s still exists", rs.Primary.ID)
			}
		}

		if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckServiceCatalogAssociateExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No product association configured")
		}

		productId := rs.Primary.Attributes["product_id"]
		portfolioId := rs.Primary.Attributes["portfolio_id"]

		conn := testAccProvider.Meta().(*AWSClient).scconn
		resp, err := conn.ListPortfoliosForProduct(&servicecatalog.ListPortfoliosForProductInput{
			ProductId: aws.String(productId),
		})
		if err != nil {
			return err
		}

		for _, p := range resp.PortfolioDetails {
			if aws.StringValue(p.Id) == portfolioId {
				return nil
			}
		}

		return fmt.Errorf("Association not found")
	}
}

func testAccAWSAssociateProductWithPortfolioImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		portfolioId := rs.Primary.Attributes["portfolio_id"]
		productId := rs.Primary.Attributes["product_id"]

		var parts []string
		parts = append(parts, portfolioId)
		parts = append(parts, productId)

		return strings.Join(parts, "_"), nil
	}
}

func testAccCheckAwsServiceCatalogResourceConfigTemplate(bucketName, portfolioName, portfolioName2, productName, provisioningArtifactName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" { }

resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  region        = "${data.aws_region.current.name}"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "template1" {
  bucket  = "${aws_s3_bucket.bucket.id}"
  key     = "test_templates_for_terraform_sc_dev1.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF template for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "arbitrary portfolio description"
  provider_name = "provider name"
}

resource "aws_servicecatalog_portfolio" "test2" {
  name          = "%s"
  description   = "arbitrary portfolio description"
  provider_name = "provider name 2"
}

resource "aws_servicecatalog_product" "test" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "%s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "%s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }
}

resource "aws_servicecatalog_associate_product_with_portfolio" "test" {
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id = aws_servicecatalog_product.test.id
}`, bucketName, portfolioName, portfolioName2, productName, provisioningArtifactName)
}

func testAccCheckAwsServiceCatalogResourceConfigUpdateTemplate(bucketName, portfolioName, portfolioName2, productName, provisioningArtifactName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" { }

resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  region        = "${data.aws_region.current.name}"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "template1" {
  bucket  = "${aws_s3_bucket.bucket.id}"
  key     = "test_templates_for_terraform_sc_dev1.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF template for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = "%s"
  description   = "arbitrary portfolio description"
  provider_name = "provider"
}

resource "aws_servicecatalog_portfolio" "test2" {
  name          = "%s"
  description   = "arbitrary portfolio description"
  provider_name = "provider 2"
}

resource "aws_servicecatalog_product" "test" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "%s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "%s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }
}

resource "aws_servicecatalog_associate_product_with_portfolio" "test" {
  portfolio_id = aws_servicecatalog_portfolio.test2.id
  product_id = aws_servicecatalog_product.test.id
}`, bucketName, portfolioName, portfolioName2, productName, provisioningArtifactName)
}
