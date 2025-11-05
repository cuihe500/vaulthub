package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var trans ut.Translator

// Init 初始化翻译器
// 必须在应用启动时调用一次
func Init() error {
	// 获取 gin 使用的 validator 实例
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return errors.New("无法获取 validator 实例")
	}

	// 注册字段名翻译函数
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// 优先使用 json tag 作为字段名
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	// 创建中文翻译器
	zhLocale := zh.New()
	uni := ut.New(zhLocale, zhLocale)
	var found bool
	trans, found = uni.GetTranslator("zh")
	if !found {
		return errors.New("无法创建中文翻译器")
	}

	// 注册中文翻译
	if err := zhTranslations.RegisterDefaultTranslations(v, trans); err != nil {
		return fmt.Errorf("注册中文翻译失败: %w", err)
	}

	return nil
}

// TranslateError 翻译 validator 错误为中文
func TranslateError(err error) string {
	if err == nil {
		return ""
	}

	// 检查是否为 validator.ValidationErrors
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		// 不是 validator 错误，直接返回原始消息
		return err.Error()
	}

	// 翻译所有错误
	var messages []string
	for _, e := range validationErrs {
		messages = append(messages, e.Translate(trans))
	}

	// 合并多个错误消息
	return strings.Join(messages, "; ")
}
