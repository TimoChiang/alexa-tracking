package main

import (
	"fmt"
	"github.com/TimoChiang/tracking"
	"github.com/arienmalec/alexa-go"
	"github.com/aws/aws-lambda-go/lambda"
	"math/rand"
	"strings"
)

// lambda handler
func Handler(request alexa.Request) (alexa.Response, error) {
	var response alexa.Response
	switch request.Body.Type {
	case "LaunchRequest": // Invoke a skill with no specific request (no intent)
		answers := []string{
			"ようこそ、宅配こです",
			"こんにちは、宅配こです",
		}
		response = NewResponse("Greeting", answers[rand.Intn(len(answers))], false)
	case "IntentRequest": // Invoke a skill with a specific request (intent)
		response =  DispatchIntents(request)
	default:
		response = NewResponse("No Body Type Catch", request.Body.Type + " is not setting", true)
	}
	return response, nil
}

func main() {
	lambda.Start(Handler)
}

// DispatchIntents dispatches each intent to the right handler
func DispatchIntents(request alexa.Request) alexa.Response {
	var response alexa.Response
	fmt.Println(request.Body.Intent.Name)
	switch request.Body.Intent.Name {
	case "TrackingRequestIntent":
		response = handleTracking(request)
	case "SetAliasIntent":
		response = handleSetAliasForTracking(request)
	case "getAllTrackingIntent":
		response = handleAllTracking(request)
	case alexa.HelpIntent:
		response = handleTrackingHelp()
	default :
		response = handleHelp()
	}

	return response
}

func handleTracking(request alexa.Request) alexa.Response {
	slots := request.Body.Intent.Slots
	company := getSlotFirstResolution(slots["trackingCompany"])
	number := slots["trackingNumberOne"].Value + slots["trackingNumberTwo"].Value + slots["trackingNumberThree"].Value + slots["trackingNumberFour"].Value
	fmt.Println(company)
	fmt.Println(number)
	if strings.Contains(number, "?") {
		return NewResponse("tracking", "ごめんなさい、番号聞き取れなかった、もう一度お願いします。", false)
	}
	if company != "" && number != "" {
		t := tracking.New()
		t.SetCompany(company)
		t.SetNumber(number)
		t.Request()
		fmt.Println(t.Result)
		fmt.Println(t.Status)
		fmt.Println(t.StatusList)
		if t.Result == "0" {
			if t.Status == "配達完了" {
				putTrackingNumber(request.Session.User.UserID, string(t.Company), t.Number, "1")
				return NewResponse("tracking", "もう配達完了しました！", true)
				//removeTrackingNumber()
			}else {
				putTrackingNumber(request.Session.User.UserID, string(t.Company), t.Number, "0")
				sessionAttribute := make(map[string]interface{})
				sessionAttribute["company"] = t.Company
				sessionAttribute["number"] = t.Number
				return NewResponseWithSession("tracking", "ただいまの状態は" + t.Status + "です。よろしければこの追跡内容を教えていただけませんか？", sessionAttribute)
				//return NewResponse("tracking", "ただいまの状態は" + t.Status + "です。", true)
			}

		} else {
			if t.Status == "伝票番号未登録" {
				putTrackingNumber(request.Session.User.UserID, string(t.Company), t.Number, "0")
				sessionAttribute := make(map[string]interface{})
				sessionAttribute["company"] = t.Company
				sessionAttribute["number"] = t.Number
				return NewResponseWithSession("tracking", "伝票番号は未登録です。反映まで少々待ちください。よろしければこの追跡内容を教えていただけませんか？", sessionAttribute)
			}else if t.Status == "伝票番号誤り" {
				return NewResponse("tracking", "すみません、この伝票番号誤ります。番号は、"+ convertNumberToKanji(number) + "間違いないでしょうか？", false)
			}
		}
	}

	return NewResponse("tracking", "すみません、よくわかりません、もう一度お願いします。", false)
}

func handleSetAliasForTracking(request alexa.Request) alexa.Response {
	sessionAttribute := request.Session.Attributes
	if len(sessionAttribute) > 0 {
		slots := request.Body.Intent.Slots
		alias := slots["tackingAlias"].Value
		putTrackingNumberWithAlias(request.Session.User.UserID, fmt.Sprintf("%v", sessionAttribute["company"]), fmt.Sprintf("%v", sessionAttribute["number"]), alias)
		return NewResponse("tracking", "設定しました！", true)
	}else{
		return NewResponse("tracking", "追跡番号まだわからないです。", false)
	}
}

func handleAllTracking(request alexa.Request) alexa.Response {
	trackings := getUserAllTrackingNumber(request.Session.User.UserID)
	message :=  fmt.Sprintf("ただいまの荷物は%d個です。\n", len(trackings))
	for k, item := range trackings {
		trackingCompositeNumber := strings.Split(item.TrackingCompositeNumber, "_")
		t := tracking.New()
		t.SetCompany(trackingCompositeNumber[0])
		t.SetNumber(trackingCompositeNumber[1])
		t.Request()

		name := ""
		if item.Alias != "" {
			name = item.Alias
		}else{
			name = convertNumberToKanji(trackingCompositeNumber[1])
		}
		message += fmt.Sprintf("%d、%sの、%s、の荷物状況は、%s。\n", k + 1, trackingCompositeNumber[0], name, t.Status)
		if t.Status == "配達完了" {
			putTrackingNumber(request.Session.User.UserID, string(t.Company), t.Number, "1")
		}
	}
	return NewResponse("All tracking ", message, true)
}

func handleTrackingHelp() alexa.Response {
	return NewResponse("Help for Tracking", "追跡サービスです", false)
}

func handleHelp() alexa.Response {
	return NewResponse("Help for Hello", "To receive a greeting, ask hello to say hello", false)
}

func NewResponse(title string, text string, isSessionEnd bool) alexa.Response {
	r := alexa.Response{
		Version: "1.0",
		Body: alexa.ResBody{
			OutputSpeech: &alexa.Payload{
				Type: "PlainText",
				Text: text,
			},
			Card: &alexa.Payload{
				Type:    "Simple",
				Title:   title,
				Content: text,
			},
			ShouldEndSession: isSessionEnd,
		},
	}
	return r
}

func NewResponseWithSession(title string, text string , sessionAttribute map[string]interface{}) alexa.Response {
	r := alexa.Response{
		Version: "1.0",
		Body: alexa.ResBody{
			OutputSpeech: &alexa.Payload{
				Type: "PlainText",
				Text: text,
			},
			Card: &alexa.Payload{
				Type:    "Simple",
				Title:   title,
				Content: text,
			},
			ShouldEndSession: false,
		},
		SessionAttributes: sessionAttribute,
	}
	return r
}

func getSlotFirstResolution(slot alexa.Slot) (resolution string) {
	resolutions := slot.Resolutions
	if len(resolutions.ResolutionPerAuthority) > 0 {
		resolution = resolutions.ResolutionPerAuthority[0].Values[0].Value.Name
	} else {
		resolution = slot.Value
	}
	return resolution
}

func convertNumberToKanji(number string) (kanji string) {
	numToKanjiMap := map[rune]string {
		'1': "一",
		'2': "二",
		'3': "三",
		'4': "四",
		'5': "五",
		'6': "六",
		'7': "七",
		'8': "八",
		'9': "九",
		'0': "ゼロ",
	}

	for _, v := range number {
		kanji += numToKanjiMap[v] + "、"
	}
	return kanji
}