package random

//const aliasLength = 4
//
//func RandomAlias(aliasLength int) string {
//	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
//
//	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "abcdefghijklmnopqrstuvwxyz" + "0123456789")
//
//	b := make([]rune, aliasLength)
//	for i := range b {
//		b[i] = chars[rnd.IntN(len(chars))]
//	}
//	return string(b)
//}
