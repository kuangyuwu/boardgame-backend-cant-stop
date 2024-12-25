package lobby

type Data struct {
	Type string `json:"type"`
	Body any    `json:"body"`
}

func dataError(errMsg string) Data {
	return Data{
		Type: "error",
		Body: errMsg,
	}
}

func dataWarn(warnMsg string) Data {
	return Data{
		Type: "warn",
		Body: warnMsg,
	}
}

func dataUsername() Data {
	return Data{
		Type: "username",
		Body: nil,
	}
}

func dataPrep() Data {
	return Data{
		Type: "prep",
		Body: nil,
	}
}

func dataUsernames(usernames []string) Data {
	return Data{
		Type: "usernames",
		Body: usernames,
	}
}
