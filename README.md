# statuspage-recurring-maintenance

A Golang command line tool to register recurring schduled maintence to [Statuspage](https://www.atlassian.com/software/statuspage).

Blog: https://nshmura.com/posts/automatic-registration-to-statuspage/

# Getting Started

## Seting up

Install [Go 1.18.x](https://go.dev/dl/)

Run below commands: 
```shell
# Install go libraries
go mod tidy

# Set API key of sutatuspage. 
export STATUSPAGE_API_KEY=.....
```

To get API Key, see [Create and manage API keys](https://support.atlassian.com/statuspage/docs/create-and-manage-api-keys/)

## Create configuration files

Create `statuspage.yaml` file like below:

```yaml
statuspagePageId: wzv88f5vctsh   # Statuspage's pageId. This values should be taken from Statuspage console.
statuspageServices:              # List of Services. Service is a group to bind multiple components of Statuspage.
  - service: ServiceA            # Service name. You can define any service name to bind some components. This name is used in `schdule.yaml` file. 
    description: Service A       # Descripton for Service (option)
    componentIds: ["pmws92dptvrm", "rww4w99psgsx"] # ComponentId of Statuspage's Component. This values should be taken from Statuspage console.

  - service: ServiceB
    ....
```

Create `schdule.yaml` file, like below:

```yaml
- service: ServiceA                            # Service name. Defined in `statuspage.yaml` file
  title: "Maintenance of ServiceA"             # Title of Scheduled Maintenance in Statuspage
  body: "... {Description for mantenance} ..." # Description of Scheduled Maintenance in Statuspage
  recurring:
    - day: everyday       # everyday maintenance
      start: 10h05m       # start time of maintenance
      time:  20m          # duration of maintenance

    - day: 2nd saturday   # 2nd saturday in every month
      start: 20h00m
      time: 10h00m

- service: ServiceB
    ....

```


## Dry run

Below example will show plans to register 3 days schduled maintence from 2023-01-01.

```
$ go run main.go recurring \
  -schedule config/schedule.yaml \
  -statuspage config/statuspage.yaml \
  -from 2023-01-01 \
  -day 3 \
  -dryRun

[dryRun]: add [Maintenance of ServiceA] 2023-01-01 10:05:00 +0900 JST - 2023-01-01 10:25:00 +0900 JST
[dryRun]: add [Maintenance of ServiceA] 2023-01-02 10:05:00 +0900 JST - 2023-01-02 10:25:00 +0900 JST
[dryRun]: add [Maintenance of ServiceA] 2023-01-03 10:05:00 +0900 JST - 2023-01-03 10:25:00 +0900 JST
```

## Register

Below example will register 3 days schduled maintence from 2023-01-01.

```
$ go run main.go recurring \
  --schdule config/schdule.yaml \
  --statuspage config/statuspage.yaml \
  --from 2023-01-01 \
  --day 3

add [Maintenance of ServiceA] 2023-01-01 10:05:00 +0900 JST - 2023-01-01 10:25:00 +0900 JST
add [Maintenance of ServiceA] 2023-01-02 10:05:00 +0900 JST - 2023-01-02 10:25:00 +0900 JST
add [Maintenance of ServiceA] 2023-01-03 10:05:00 +0900 JST - 2023-01-03 10:25:00 +0900 JST
```


# Command Options

```
% go run main.go -h
Usage:
  -day int
    	days of terms to create schedule
  -dryRun
    	is dryRun
  -from string
    	first date to create schedule
  -schdule string
    	file to load configuration of schdule for maintenance
  -statuspage string
    	file to load configuration of statuspage
```
