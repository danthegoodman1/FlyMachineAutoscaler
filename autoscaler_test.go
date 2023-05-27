package main

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	"testing"
	"time"
)

func TestPromQuery1(t *testing.T) {
	godotenv.Load()

	res, err := QueryPrometheus(context.Background(), "(sum(rate(fly_instance_cpu{mode!=\"idle\"}[60s]))by(mode)>0) / count(fly_instance_cpu{mode=\"idle\"})", time.UnixMilli(1685195160000))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(spew.Sdump(res))
}

func TestPromQuery2(t *testing.T) {
	godotenv.Load()

	res, err := QueryPrometheus(context.Background(), "(avg((sum(rate(fly_instance_cpu{mode!=\"idle\"}[60s]))by(region, instance)>0)) by (region)) / count(fly_instance_cpu{mode=\"idle\"})", time.Unix(1685196210, 0))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(spew.Sdump(res))
	//	{"status":"success","isPartial":false,"data":{"resultType":"vector","result":[{"metric":{"region":"ams"},"value":[1685196210,"0.01666722224074136"]},{"metric":{"region":"arn"},"value":[1685196210,"0.019270833333333334"]},{"metric":{"region":"atl"},"value":[1685196210,"0.013541666666666667"]},{"metric":{"region":"bog"},"value":[1685196210,"0.015625"]},{"metric":{"region":"bos"},"value":[1685196210,"0.014583333333333334"]},{"metric":{"region":"cdg"},"value":[1685196210,"0.011979166666666666"]},{"metric":{"region":"den"},"value":[1685196210,"0.014583333333333334"]},{"metric":{"region":"dfw"},"value":[1685196210,"0.030730191006366883"]},{"metric":{"region":"ewr"},"value":[1685196210,"0.18541666666666667"]},{"metric":{"region":"eze"},"value":[1685196210,"0.0125"]},{"metric":{"region":"gdl"},"value":[1685196210,"0.0078125"]},{"metric":{"region":"gig"},"value":[1685196210,"0.01640625"]},{"metric":{"region":"gru"},"value":[1685196210,"0.028644878504049866"]},{"metric":{"region":"hkg"},"value":[1685196210,"0.024479166666666666"]},{"metric":{"region":"iad"},"value":[1685196210,"0.014583333333333334"]},{"metric":{"region":"jnb"},"value":[1685196210,"0.034374999999999996"]},{"metric":{"region":"lax"},"value":[1685196210,"0.010937135428819038"]},{"metric":{"region":"lhr"},"value":[1685196210,"0.01171875"]},{"metric":{"region":"mad"},"value":[1685196210,"0.0203125"]},{"metric":{"region":"mia"},"value":[1685196210,"0.0203125"]},{"metric":{"region":"nrt"},"value":[1685196210,"0.030729166666666665"]},{"metric":{"region":"ord"},"value":[1685196210,"0.007031250000000001"]},{"metric":{"region":"otp"},"value":[1685196210,"0.019791666666666666"]},{"metric":{"region":"qro"},"value":[1685196210,"0.008854461815393847"]},{"metric":{"region":"scl"},"value":[1685196210,"0.011979166666666667"]},{"metric":{"region":"sea"},"value":[1685196210,"0.01875"]},{"metric":{"region":"sin"},"value":[1685196210,"0.013020833333333334"]},{"metric":{"region":"sjc"},"value":[1685196210,"0.012499999999999999"]},{"metric":{"region":"syd"},"value":[1685196210,"0.009895833333333333"]},{"metric":{"region":"waw"},"value":[1685196210,"0.023437499999999997"]},{"metric":{"region":"yul"},"value":[1685196210,"0.01614583333333333"]},{"metric":{"region":"yyz"},"value":[1685196210,"0.029686510449651676"]}]}}
}
