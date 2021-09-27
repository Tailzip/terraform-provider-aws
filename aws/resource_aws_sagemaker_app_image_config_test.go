package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_app_image_config", &resource.Sweeper{
		Name: "aws_sagemaker_app_image_config",
		F:    testSweepSagemakerAppImageConfigs,
	})
}

func testSweepSagemakerAppImageConfigs(region string) error {
	client, err := acctest.SharedRegionalSweeperClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).SageMakerConn
	input := &sagemaker.ListAppImageConfigsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListAppImageConfigs(input)

		if acctest.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker App Image Config for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Example Thing: %w", err))
			return sweeperErrs
		}

		for _, config := range output.AppImageConfigs {

			name := aws.StringValue(config.AppImageConfigName)
			r := ResourceAppImageConfig()
			d := r.Data(nil)
			d.SetId(name)
			err = r.Delete(d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SageMaker App Image Config (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSagemakerAppImageConfig_basic(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("app-image-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerAppImageConfig_kernelGatewayImageConfig_kernalSpecs(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigKernalSpecs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigKernalSpecs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.display_name", rName),
				),
			},
		},
	})
}

func TestAccAWSSagemakerAppImageConfig_kernelGatewayImageConfig_fileSystemConfig(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigFileSystemConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_gid", "100"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_uid", "1000"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.mount_path", "/home/sagemaker-user"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigFileSystemConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_gid", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_uid", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.mount_path", "/test"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerAppImageConfig_tags(t *testing.T) {
	var app sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppConfigImageConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerAppConfigImageConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerAppConfigImageConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerAppImageConfig_disappears(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceAppImageConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerAppImageConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_app_image_config" {
			continue
		}

		config, err := finder.AppImageConfigByName(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Sagemaker App Image Config (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(config.AppImageConfigName) == rs.Primary.ID {
			return fmt.Errorf("Sagemaker App Image Config %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerAppImageConfigExists(n string, config *sagemaker.DescribeAppImageConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker App Image Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := finder.AppImageConfigByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*config = *resp

		return nil
	}
}

func testAccAWSSagemakerAppImageConfigBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigKernalSpecs1(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }
  }
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigKernalSpecs2(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name         = "%[1]s-2"
      display_name = %[1]q
    }
  }
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigFileSystemConfig1(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }

    file_system_config {}
  }
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigFileSystemConfig2(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }

    file_system_config {
      default_gid = 0
      default_uid = 0
      mount_path  = "/test"
    }
  }
}
`, rName)
}

func testAccAWSSagemakerAppConfigImageConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerAppConfigImageConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
