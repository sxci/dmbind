package enum

var Cdn = struct{Wangsu,Tongxing,Lanxun int}{
	Wangsu: 0,
	Tongxing: 1,
	Lanxun: 2,
}

var Status = struct{Init,Deny,Success,Wait,ComeIn int} {
	Init: 0, Deny: 1, Success:2, Wait: 3, ComeIn: 4,
}
