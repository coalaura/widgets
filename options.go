package main

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/mazznoer/csscolorparser"
)

type Options map[string]Option

type Option interface {
	Type() string
	Default() any
	Description() string
	Value(string) any
}

type Parser func(string, string) any

type BaseOption struct {
	typ         string
	description string
}

type OptionInt struct {
	BaseOption
	def int
}

type OptionFloat struct {
	BaseOption
	def float64
}

type OptionString struct {
	BaseOption
	def   string
	parse Parser
}

type OptionEnum struct {
	BaseOption
	def     string
	allowed []string
}

func NewInt(def int, description string) Option {
	return &OptionInt{
		BaseOption: BaseOption{"number", description},
		def:        def,
	}
}

func NewFloat(def float64, description string) Option {
	return &OptionFloat{
		BaseOption: BaseOption{"number", description},
		def:        def,
	}
}

func NewString(def string, description string) Option {
	return &OptionString{
		BaseOption: BaseOption{"text", description},
		def:        def,
	}
}

func NewColor(def string, description string) Option {
	return &OptionString{
		BaseOption: BaseOption{"color", description},
		def:        def,
		parse:      ColorValue,
	}
}

func NewSize(def string, description string) Option {
	return &OptionString{
		BaseOption: BaseOption{"text", description},
		def:        def,
		parse:      SizeValue,
	}
}

func NewEnum(def string, allowed []string, description string) Option {
	return &OptionEnum{
		BaseOption: BaseOption{"select", description},
		def:        def,
		allowed:    allowed,
	}
}

func (o *Options) RegisterDefault(name, def, description string) {
	if _, ok := (*o)[name]; ok {
		return
	}

	(*o)[name] = NewString(def, description)
}

func (o Options) MarshalJSON() ([]byte, error) {
	output := make(map[string]any)

	for key, opt := range o {
		entry := map[string]any{
			"type":        opt.Type(),
			"default":     opt.Default(),
			"description": opt.Description(),
		}

		if enum, ok := opt.(*OptionEnum); ok {
			entry["allowed"] = enum.allowed
		}

		output[key] = entry
	}

	return json.Marshal(output)
}

func (o *BaseOption) Type() string {
	return o.typ
}

func (o *OptionInt) Default() any {
	return o.def
}

func (o *OptionFloat) Default() any {
	return o.def
}

func (o *OptionString) Default() any {
	return o.def
}

func (o *OptionEnum) Default() any {
	return o.def
}

func (o *BaseOption) Description() string {
	return o.description
}

func (o *OptionInt) Value(input string) any {
	if value, err := strconv.ParseInt(input, 10, 64); err == nil {
		return int(value)
	}

	return o.def
}

func (o *OptionFloat) Value(input string) any {
	if value, err := strconv.ParseFloat(input, 64); err == nil {
		return value
	}

	return o.def
}

func (o *OptionString) Value(input string) any {
	if input == "" {
		return o.def
	}

	if o.parse == nil {
		return input
	}

	return o.parse(o.def, input)
}

func (o *OptionEnum) Value(input string) any {
	if input == "" || len(o.allowed) == 0 {
		return o.def
	}

	for _, value := range o.allowed {
		if value == input {
			return value
		}
	}

	return o.def
}

func ColorValue(def, input string) any {
	col, err := csscolorparser.Parse(input)
	if err != nil {
		return def
	}

	return col.HexString()
}

func SizeValue(def, input string) any {
	// no leading zero (.25rem)
	if strings.HasPrefix(input, ".") {
		input = "0" + input
	}

	rgx := regexp.MustCompile(`(?m)^\d+(\.\d+)?(p[xtc]|r?em|in|ex|ch|[cm]m|v([wh]|m(in|ax))|%)$`)

	if !rgx.MatchString(input) {
		return def
	}

	return input
}
