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
	// once Run starts. Updates over the limit are dropped. Updates received
	// while they are replayed are buffered too, so they are handled in order.
	//
	// If Run fetches the state from the server (AuthOptions.Forget, or no
	// stored state), the buffered updates are already covered by it, so they
	// are passed to Handler as-is, like with BeforeRunPass.
	//
	// Use it when handlers are ready only after authorization, e.g.
	// UpdateDispatcher handlers registered in Run's OnStart.
	BeforeRunBuffer
	// BeforeRunDrop drops updates received before Run.
	BeforeRunDrop
)

// defaultBeforeRunBufferSize is the default Config.BeforeRunBufferSize.
const defaultBeforeRunBufferSize = 100
