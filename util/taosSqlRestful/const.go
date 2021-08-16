package taossqlrestful

const (
	timeFormat     = "2006-01-02 15:04:05"
	maxTaosSqlLen  = 65380
	defaultBufSize = maxTaosSqlLen + 32
)

type fieldType byte

type fieldFlag uint16

const (
	flagNotNULL fieldFlag = 1 << iota
)

type statusFlag uint16

const (
	statusInTrans statusFlag = 1 << iota
	statusInAutocommit
	statusReserved // Not in documentation
	statusMoreResultsExists
	statusNoGoodIndexUsed
	statusNoIndexUsed
	statusCursorExists
	statusLastRowSent
	statusDbDropped
	statusNoBackslashEscapes
	statusMetadataChanged
	statusQueryWasSlow
	statusPsOutParams
	statusInTransReadonly
	statusSessionStateChanged
)
