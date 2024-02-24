package proto

// TypesMap returns mapping from type ids to TL type names.
func TypesMap() map[uint32]string {
	return map[uint32]string{
		MessageContainerTypeID: "message_container",
		ResultTypeID:           "rpc_result",
		GZIPTypeID:             "gzip",
	}
}
