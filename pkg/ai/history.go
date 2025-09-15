package ai

type History []Message

func NewHistory(messages ...Message) History {
	return messages
}

func (h History) Append(messages ...Message) History {
	return append(h, messages...)
}

func (h History) Last() Message {
	if len(h) == 0 {
		return nil
	}

	return h[len(h)-1]
}
