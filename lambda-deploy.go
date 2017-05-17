package main

import (
	"encoding/json"
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/jhoonb/archivex"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Environments is the format in which the config is given to the app
type Environments struct {
	Dev        Config `json:"dev"`
	QA         Config `json:"qa"`
	Production Config `json:"prod"`
}

// Config is the Lambda configuration for each environment
type Config struct {
	Lambda_Name      string
	Lambda_Directory string
	Region		 string
}

func main() {
	// The directory flag passed in in the cli
	directoryPtr := flag.String("directory", "", "The path to the lambda base directory")
	// The environment flag. Should be QA, Dev or Prod
	environmentPtr := flag.String("env", "", "The environment to deploy the lambda")
	flag.Parse()

	if *directoryPtr == "" {
		log.Fatal("No directory path supplied")
	}

	if *environmentPtr == "" {
		log.Fatal("No environment supplied")
	}

	// formatting the environment
	env := strings.ToLower(*environmentPtr)

	conf := ReadConfigFile(*directoryPtr, env)
	sess := BuildAwsSession(conf)
	zip := ZipContents(conf)
	PushLambda(sess, conf, zip)
}

// ReadConfigFile accepts a path and an environment
// The path should be to the root of the lambda directory, featuring a src and a deploy.json file
// The environment relates to which environment we'd like to use to deploy, in the deploy.json file
func ReadConfigFile(path string, env string) *Config {
	log.Print("Reading Config")
	b, err := ioutil.ReadFile(path + "/deploy.json") // This could be prettier
	if err != nil {
		log.Fatal("Cannot read config file")
	}

	var e Environments
	err = json.Unmarshal(b, &e)
	if err != nil {
		log.Fatal("Cannot unmarshal config")
	}

	var c Config
	switch env { // This could be prettier
	case "dev":
		c = e.Dev
	case "qa":
		c = e.QA
	case "prod":
		c = e.Production
	}

	if c.Region == "" {
		c.Region = "us-west-2"
	}

	c.Lambda_Directory = path // The Lambda Directory is always the same
	return &c
}

// BuildAwsSession accepts an environment configuration and returns an AWS Session pointer
// To be able to create a session it is required that your machine has an appropriate AWS credentials file at
// ~/.aws/credentials
func BuildAwsSession(c *Config) *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(c.Region),
	})

	if err != nil {
		log.Print(err)
		log.Fatal("Cannot create AWS Session")
	}

	return sess
}

// ZipContents zips the contents of a config-specified directory
// It returns the location of that zipped file
func ZipContents(c *Config) string {
	srcLocation := c.Lambda_Directory + "/src"
	zipLocation := c.Lambda_Name + ".zip"

	zipFile := new(archivex.ZipFile)
	zipFile.Create(zipLocation)
	log.Print("Attempting to Add Directory")

	err := zipFile.AddAll(srcLocation, false)
	if err != nil {
		log.Print(err)
		log.Fatal("Cannot create zip")
	}

	zipFile.Close()
	log.Print("Directory Added")
	return zipLocation
}

// PushLambda pushes a zip file to the provided lambda.
// This cannot guarantee the contents of the lambda are correct, but will push it to the appropriate environment
func PushLambda(sess *session.Session, c *Config, zip string) {
	svc := lambda.New(sess, &aws.Config{
		Region: aws.String(c.Region),
	})

	contents, err := ioutil.ReadFile(zip)
	if err != nil {
		log.Print(err)
		log.Fatal("Coulnd't upload Lambda Zip")
	}
	log.Print("Attempting to Push Code")
	_, err = svc.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(c.Lambda_Name),
		ZipFile:      contents,
		Publish:      aws.Bool(true),
	})

	if err != nil {
		os.Remove(zip)
		log.Print(err)
		log.Fatal("Invalid Lambda Input Config")
	} else {
		log.Print("Code Pushed")
		os.Remove(zip)
	}
}
