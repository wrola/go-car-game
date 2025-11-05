package game

type KeyMapping struct {
	UpKey    string
	DownKey  string
	LeftKey  string
	RightKey string
}

var (
	WASDMapping = KeyMapping{
		UpKey:    "w",
		DownKey:  "s",
		LeftKey:  "a",
		RightKey: "d",
	}

	ArrowKeyMapping = KeyMapping{
		UpKey:    "up",
		DownKey:  "down",
		LeftKey:  "left",
		RightKey: "right",
	}
)

type InputHandler struct {
	keyMapping KeyMapping
}

func NewInputHandler(keyMapping KeyMapping) *InputHandler {
	return &InputHandler{
		keyMapping: keyMapping,
	}
}

func (ih *InputHandler) Translate(rawInput map[string]interface{}) map[string]bool {
	result := make(map[string]bool)
	if val, ok := rawInput[ih.keyMapping.UpKey].(bool); ok {
		result["up"] = val
	}
	if val, ok := rawInput[ih.keyMapping.DownKey].(bool); ok {
		result["down"] = val
	}
	if val, ok := rawInput[ih.keyMapping.LeftKey].(bool); ok {
		result["left"] = val
	}
	if val, ok := rawInput[ih.keyMapping.RightKey].(bool); ok {
		result["right"] = val
	}

	return result
}
