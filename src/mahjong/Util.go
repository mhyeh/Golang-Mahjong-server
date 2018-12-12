package mahjong

// IF implements ternary conditional operator
func IF(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}