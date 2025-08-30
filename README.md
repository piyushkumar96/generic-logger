# Generic Logger
This is generic logger used for logging the application data. Currently it supports zap-logger as base logger.

## Description
1. Zap logger is used as base logger because of following 
   
   1. Zap is an open-source, fast, structured, and efficient logging library for Go and also provides custom log formats, custom log levels, sampling, error wrapping, and more. 
   2. It provides a simple API for emitting structured and leveled logs, that can be consumed by various log aggregators or output sinks, such as ElasticSearch, Kafka, or stdout. It is designed to be very performant, memory-efficient and low-latency, making it suitable for high-throughput production environments. 
   3. It is used by many companies and organizations, including Uber, Lyft, Square, Docker, Netflix, and many others.

2. Currently, it supports **console**, **file logging** and **socket logging** with help of vector.

## Usage

```shell
   $ go get github.com/piyushkumar96/generic-logger@Version   
```

>Note: Export following env variable to enable/disable specific feature of logger

```shell
   LOGGER_MODE: INFO -- to set logging level support value DEBUG, ERROR, INFO(default)
   LOGGER_JSON_ENCODER_DISABLED: true -- to disable the json encoding of logs
   LOGGER_CONSOLE_SYNCER_DISABLED: true -- to disable the std out based logging of logs
   LOGGER_FILE_SYNCER_DISABLED: true -- to disable file based logging of logs
   LOGGER_SOCKET_LOGGING_ENABLED: true -- to enable socket logging of logs
```

```go
   package main
   import (
	   "github.com/piyushkumar96/generic-logger"
   )

   func main()  {
      logger.Info("successfully run the validation rule", "code", "OnValidationRuleExecutionSuccess", "meta", map[string]interface{}{"rule_name": "ImpressionsRateDrop>30", "validation_type": "Spot", "alert_level": "SupplyTag", "aggregation_level": "hourly"})
	   var err error
	   if err != nil {
          logger.Error("failed to run the validation rule", "code", "OnValidationRuleExecutionFailure","err", err.Error(), "meta", map[string]interface{}{"rule_name": "ImpressionsRateDrop>30", "validation_type": "Spot", "alert_level": "SupplyTag", "aggregation_level": "hourly"})
       }
   }
```


---

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

---

## ⭐ Show Your Support

If this repository helped you learn software engineering concepts, please:
- ⭐ **Star** this repository
- 🍴 **Fork** it for your own learning
- 📢 **Share** it with others who might benefit
- 💬 **Provide feedback** through issues or discussions

---

*Transform your Python skills and software engineering knowledge with these comprehensive, practical examples! Start your journey to writing cleaner, more maintainable code today! 🚀*

**Happy Learning and Happy Coding! 🐍✨** 