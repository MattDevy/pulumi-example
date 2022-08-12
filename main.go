package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	shell     = "sh"
	shellFlag = "-c"
)

var rootFolder string

func init() {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	rootFolder = path
}

func runCmd(args string) error {
	cmd := exec.Command(shell, shellFlag, args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = rootFolder
	return cmd.Run()
}

func bundleLambda() error {
	if err := runCmd("GOOS=linux GOARCH=amd64 go build -o ./bin/hello-world ./cmd/hello-world"); err != nil {
		return err
	}

	if err := runCmd("zip -r -j ./bin/hello-world.zip ./bin/hello-world"); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := bundleLambda(); err != nil {
		log.Fatalf("failed to bundle lambda: %v", err)
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an IAM role.
		role, err := iam.NewRole(ctx, "task-exec-role", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Sid": "",
					"Effect": "Allow",
					"Principal": {
						"Service": "lambda.amazonaws.com"
					},
					"Action": "sts:AssumeRole"
				}]
			}`),
		})
		if err != nil {
			return err
		}

		// Attach a policy to allow writing logs to CloudWatch
		logPolicy, err := iam.NewRolePolicy(ctx, "lambda-log-policy", &iam.RolePolicyArgs{
			Role: role.Name,
			Policy: pulumi.String(`{
                "Version": "2012-10-17",
                "Statement": [{
                    "Effect": "Allow",
                    "Action": [
                        "logs:CreateLogGroup",
                        "logs:CreateLogStream",
                        "logs:PutLogEvents"
                    ],
                    "Resource": "arn:aws:logs:*:*:*"
                }]
            }`),
		})

		// Set arguments for constructing the function resource.
		args := &lambda.FunctionArgs{
			Runtime: lambda.RuntimeGo1dx,
			Role:    role.Arn,
			Code:    pulumi.NewFileArchive("./bin/hello-world.zip"),
			Handler: pulumi.String("hello-world"),
		}

		// Create the lambda using the args.
		function, err := lambda.NewFunction(
			ctx,
			"basicLambda",
			args,
			pulumi.DependsOn([]pulumi.Resource{logPolicy}),
		)
		if err != nil {
			return err
		}

		// Export the lambda ARN.
		ctx.Export("lambda", function.Arn)

		return nil
	})
}
