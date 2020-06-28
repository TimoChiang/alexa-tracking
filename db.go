package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var svc *dynamodb.DynamoDB
const (
	tableName = "Tracking"
	SampleId = "amzn1.ask.account.AG6QGBMNANG6CCAV2W6OOUQRKPGO7G23HQN3E3GIQGF3IOTQC3KSCUHEWFWROE653NIODBO5S2UBMFISKIXSK5GCAG55JAVNNIE5DRCDMBEJDORPKOAFK6TIXGYFFACXJFSVHV3VBW4N636BW4V424I77HACRS7L3DILQSAXYIOE7IUEHCK44DO3EXWWWE2JIK7QZD53WFWPLTI"
)

type Item struct {
	UserId         string
	TrackingCompositeNumber string //company + tracking number
	Alias string
}

func dbInitial() {
	if svc == nil {
		sess := session.Must(session.NewSession())
		svc = dynamodb.New(sess)
	}
}

func getItem(partitionKeyValue , sortKeyValue string) *Item {
	dbInitial()
	item := new(Item)
	getParams, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UserId": {
				S: aws.String(partitionKeyValue),
			},
			"TrackingCompositeNumber": {
				S: aws.String(sortKeyValue),
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return item
	}

	err = dynamodbattribute.UnmarshalMap(getParams.Item, &item)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal Record, %v", err)
		fmt.Println(errMsg)
		panic(errMsg)
	}
	return item
}

func putItem(partitionKeyValue , sortKeyValue, flg, alias string) (bool, error) {
	dbInitial()
	_, err := svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"UserId": {
				S: aws.String(partitionKeyValue),
			},
			"TrackingCompositeNumber": {
				S: aws.String(sortKeyValue),
			},
			"UpdatedAt": {
				S: aws.String("2020-06-20"),
			},
			"ShippingFinishFlg": {
				S: aws.String(flg),
			},
			"Alias": {
				S: aws.String(alias),
			},
		},
	})

	if err != nil {

		return false, err
	}

	return true, nil
}

func queryItems(partitionKeyValue , sortKeyValue, flg string) []*Item {
	dbInitial()
	items := make([]*Item, 0)
	getQuerys, err := svc.Query(&dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditionExpression: aws.String("#ID = :partitionkeyval AND begins_with ( #SORTKEY, :sortkeyval )"),
		FilterExpression : aws.String("#FALG = :flagValue "),
		ExpressionAttributeNames: map[string]*string{
			"#ID": aws.String("UserId"),
			"#SORTKEY": aws.String("TrackingCompositeNumber"),
			"#FALG": aws.String("ShippingFinishFlg"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":partitionkeyval": {
				S: aws.String(partitionKeyValue),
			},
			":sortkeyval": {
				S: aws.String(sortKeyValue),
			},
			":flagValue": {
				S: aws.String(flg),
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return items
	}

	dynamodbattribute.UnmarshalListOfMaps(getQuerys.Items, &items)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal Record, %v", err)
		fmt.Println(errMsg)
		panic(errMsg)
	}
	return items
}

func queryItemsWithRangeKey(partitionKeyValue , sortKeyValue string) []*Item {
	dbInitial()
	items := make([]*Item, 0)
	getQuerys, err := svc.Query(&dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditionExpression: aws.String("#ID = :partitionkeyval AND begins_with ( #SORTKEY, :sortkeyval )"),
		ExpressionAttributeNames: map[string]*string{
			"#ID": aws.String("UserId"),
			"#SORTKEY": aws.String("TrackingCompositeNumber"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":partitionkeyval": {
				S: aws.String(partitionKeyValue),
			},
			":sortkeyval": {
				S: aws.String(sortKeyValue),
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return items
	}

	dynamodbattribute.UnmarshalListOfMaps(getQuerys.Items, &items)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal Record, %v", err)
		fmt.Println(errMsg)
		panic(errMsg)
	}
	return items
}

func queryItemsWithPrimaryKey(partitionKeyValue string) []*Item {
	dbInitial()
	items := make([]*Item, 0)
	getQuerys, err := svc.Query(&dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditionExpression: aws.String("#ID = :partitionkeyval"),
		FilterExpression : aws.String("#FALG = :flagValue "),
		ExpressionAttributeNames: map[string]*string{
			"#ID": aws.String("UserId"),
			"#FALG": aws.String("ShippingFinishFlg"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":partitionkeyval": {
				S: aws.String(partitionKeyValue),
			},
			":flagValue": {
				S: aws.String("0"),
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return items
	}

	dynamodbattribute.UnmarshalListOfMaps(getQuerys.Items, &items)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal Record, %v", err)
		fmt.Println(errMsg)
		panic(errMsg)
	}
	return items
}

func deleteItem(partitionKeyValue , sortKeyValue string) {
	dbInitial()
	_, err := svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UserId": {
				S: aws.String(partitionKeyValue),
			},
			"TrackingCompositeNumber": {
				S: aws.String(sortKeyValue),
			},
		},
	})

	if err != nil {
		fmt.Println("Got error calling DeleteItem")
		fmt.Println(err.Error())
		return
	}
}

func getTrackingNumber(userId, company, num string) *Item {
	return getItem(userId, getTrackingCompositeNumber(company, num))
}

func putTrackingNumber(userId, company, num, flg string) {
	if ok, err := putItem(userId, getTrackingCompositeNumber(company, num), flg, ""); !ok {
		panic(err)
	}
}

func putTrackingNumberWithAlias(userId, company, num, alias string) {
	if ok, err := putItem(userId, getTrackingCompositeNumber(company, num), "0", alias); !ok {
		panic(err)
	}
}

func getTrackingNumberWithCompany(userId, company string) []*Item {
	return queryItemsWithRangeKey(userId, company + "_")
}

func getAvailableTrackingNumber(userId, company, num string) []*Item {
	return queryItems(userId, getTrackingCompositeNumber(company, num), "0")
}

func removeTrackingNumber(userId, company, num string) {
	deleteItem(userId, getTrackingCompositeNumber(company, num))
}


func getTrackingCompositeNumber(company, num string) string {
	return company + "_" + num
}

func getUserAllTrackingNumber(userId string) []*Item {
	return queryItemsWithPrimaryKey(userId)
}
