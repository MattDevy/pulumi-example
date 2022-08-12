# pulumi-example

## Prerequisites
- Install pulumi 
```sh
brew install pulumi
```
- Install Go
- Set up AWS profiles

## Get started
```sh
# clone this repo 
git clone github.com/MattDevy/pulumi-example

# login to pulumi (locally for demo)
# this should create a stack also
pulumi login --local # (follow the steps)

# if stack not created, do
pulumi stack init

# set aws profile
pulumi config set aws:profile dev-hydra-ecs

# create the example, press yes when time comes
pulumi up

# invoke the lambda
aws lambda invoke \
--function-name $(pulumi stack output lambda) \
--region $(pulumi config get aws:region) \
--cli-binary-format raw-in-base64-out output.json

# see response
cat output.json
```

## Tidy up
```sh
pulumi destroy --yes
pulumi stack rm --yes
```
