package apache

import (
	"github.com/turbot/tailpipe-plugin-apache/tables/access_log"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const PluginName = "apache"

func init() {
	// Register the table, with type parameter:
	// 1. table type
	table.RegisterCustomTable[*access_log.AccessLogTable]()

	// register formats
	table.RegisterFormat[*access_log.AccessLogTableFormat]()
	table.RegisterFormatPresets(access_log.AccessLogTableFormatPresets...)
}

type Plugin struct {
	plugin.PluginImpl
}

func NewPlugin() (_ plugin.TailpipePlugin, err error) {
	p := &Plugin{
		PluginImpl: plugin.NewPluginImpl(PluginName),
	}

	return p, nil
}
