package task_update

// Optional describes a partial-update field.
// Set=false means the field was omitted; Set=true with Valid=false means explicit null.
type Optional[T any] struct {
	Set   bool
	Valid bool
	Value T
}
