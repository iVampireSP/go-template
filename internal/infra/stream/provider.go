package stream

// 按需使用 MQTT 或 Kafka

import (
	"github.com/eclipse/paho.mqtt.golang"
	"go-template/internal/infra/conf"
)

type Stream struct {
	MQTT     mqtt.Client
	MQTTOpts mqtt.ClientOptions
}

func NewStream(config *conf.Config) *Stream {
	return nil
	//opts := mqtt.NewClientOptions()
	//
	//// default hostname
	//host, err := os.Hostname()
	//if err != nil {
	//	panic(err)
	//}
	//
	//if config.Stream.MQTTClientId == "" {
	//	opts.SetClientID(host)
	//} else {
	//	opts.SetClientID(config.Stream.MQTTClientId + "-" + host)
	//}
	//
	//for _, broker := range config.Stream.MQTTBrokers {
	//	opts.AddBroker(broker)
	//}
	//
	//opts.SetPassword(config.Stream.MQTTPassword)
	//opts.SetUsername(config.Stream.MQTTUsername)
	//
	//opts.SetKeepAlive(60 * time.Second)
	//
	//opts.SetPingTimeout(5 * time.Second)
	//
	//c := mqtt.NewClient(opts)
	//
	//if token := c.Connect(); token.Wait() && token.Error() != nil {
	//	panic("Can not connect to MQTT:" + token.Error().Error())
	//}
	//
	//streamObj := &Stream{
	//	MQTT:     c,
	//	MQTTOpts: *opts,
	//}
	//
	//return streamObj
}
