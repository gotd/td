package updates

// BeforeRunMode is what Manager does with updates received before Run.
type BeforeRunMode int

const (
	// BeforeRunPass passes updates received before Run to Handler as-is,
	// without order verification, gap recovery or short updates
	// transformation. Default.
	BeforeRunPass BeforeRunMode = iota
	// BeforeRunBuffer keeps updates received before Run, at most
	// Config.BeforeRunBufferSize batches, and pushes them through gap recovery
	// once Run starts. Updates over the limit are dropped.
	//
	// Use it when handlers are ready only after authorization, e.g.
	// UpdateDispatcher handlers registered in Run's OnStart.
	BeforeRunBuffer
	// BeforeRunDrop drops updates received before Run.
	BeforeRunDrop
)

// defaultBeforeRunBufferSize is the default Config.BeforeRunBufferSize.
const defaultBeforeRunBufferSize = 100
