package mgl64

import (
	"math"
)

// A rotation order is the order in which
// rotations may be transformed for the purposes of AnglesToQuat
type RotationOrder int

const (
	XYX RotationOrder = iota
	XYZ
	XZX
	XZY
	YXY
	YXZ
	YZY
	YZX
	ZYZ
	ZYX
	ZXZ
	ZXY
)

type Quat struct {
	W float64
	V Vec3
}

func QuatIdent() Quat {
	return Quat{1., Vec3{0, 0, 0}}
}

func QuatRotate(angle float64, axis Vec3) Quat {
	angle = (float64(math.Pi) * angle) / 180.0

	c, s := float64(math.Cos(float64(angle/2))), float64(math.Sin(float64(angle/2)))

	return Quat{c, axis.Mul(s)}
}

// This function is deprecated. Instead, use AnglesToQuat
//
// The behavior of this function should be equivalent to AnglesToQuat(zAngle, yAngle, xAngle, ZYX)
func EulerToQuat(xAngle, yAngle, zAngle float64) Quat {
	sinz, cosz := math.Sincos(float64(zAngle))
	sz, cz := float64(sinz), float64(cosz)

	siny, cosy := math.Sincos(float64(yAngle))
	sy, cy := float64(siny), float64(cosy)

	sinx, cosx := math.Sincos(float64(xAngle))
	sx, cx := float64(sinx), float64(cosx)

	return Quat{
		W: cx*cy*cz + sx*sy*sz,
		V: Vec3{
			sx*cy*cz - cx*sy*sz,
			cx*sy*cz + sx*cy*sz,
			cx*cy*sz - sx*sy*cz,
		},
	}
}

func (q Quat) X() float64 {
	return q.V[0]
}

func (q Quat) Y() float64 {
	return q.V[1]
}

func (q Quat) Z() float64 {
	return q.V[2]
}

func (q1 Quat) Add(q2 Quat) Quat {
	return Quat{q1.W + q2.W, q1.V.Add(q2.V)}
}

func (q1 Quat) Sub(q2 Quat) Quat {
	return Quat{q1.W - q2.W, q1.V.Sub(q2.V)}
}

func (q1 Quat) Mul(q2 Quat) Quat {
	return Quat{q1.W*q2.W - q1.V.Dot(q2.V), q1.V.Cross(q2.V).Add(q2.V.Mul(q1.W)).Add(q1.V.Mul(q2.W))}
}

func (q1 Quat) Scale(c float64) Quat {
	return Quat{q1.W * c, Vec3{q1.V[0] * c, q1.V[1] * c, q1.V[2] * c}}
}

func (q1 Quat) Conjugate() Quat {
	return Quat{q1.W, q1.V.Mul(-1)}
}

func (q1 Quat) Len() float64 {
	return float64(math.Sqrt(float64(q1.W*q1.W + q1.V[0]*q1.V[0] + q1.V[1]*q1.V[1] + q1.V[2]*q1.V[2])))
}

func (q1 Quat) Normalize() Quat {
	length := q1.Len()

	if FloatEqual(1, length) {
		return q1
	}

	return Quat{q1.W * 1 / length, q1.V.Mul(1 / length)}
}

func (q1 Quat) Inverse() Quat {
	leng := q1.Len()
	return Quat{q1.W, q1.V.Mul(-1)}.Scale(1 / (leng * leng))
}

func (q1 Quat) Rotate(v Vec3) Vec3 {
	return q1.Mul(Quat{0, v}).Mul(q1.Conjugate()).V
}

func (q1 Quat) Mat4() Mat4 {
	w, x, y, z := q1.W, q1.V[0], q1.V[1], q1.V[2]
	return Mat4{1 - 2*y*y - 2*z*z, 2*x*y + 2*w*z, 2*x*z - 2*w*y, 0, 2*x*y - 2*w*z, 1 - 2*x*x - 2*z*z, 2*y*z + 2*w*x, 0, 2*x*z + 2*w*y, 2*y*z - 2*w*x, 1 - 2*x*x - 2*y*y, 0, 0, 0, 0, 1}
}

func (q1 Quat) Dot(q2 Quat) float64 {
	return q1.W*q1.W + q1.V[0]*q1.V[0] + q1.V[1]*q1.V[1] + q1.V[2]*q1.V[2]
}

func QuatSlerp(q1, q2 Quat, amount float64) Quat {
	q1, q2 = q1.Normalize(), q2.Normalize()
	dot := q1.Dot(q2)

	// This is here for precision errors, I'm perfectly aware the *technically* the dot is bound [-1,1], but since Acos will freak out if it's not (even if it's just a liiiiitle bit over due to normal error) we need to clamp it
	dot = Clampf(dot, -1, 1)

	theta := float64(math.Acos(float64(dot))) * amount
	c, s := float64(math.Cos(float64(theta))), float64(math.Sin(float64(theta)))
	rel := q2.Sub(q1.Scale(dot)).Normalize()

	return q2.Sub(q1.Scale(c)).Add(rel.Scale(s))
}

func QuatLerp(q1, q2 Quat, amount float64) Quat {
	return q1.Add(q2.Sub(q1).Scale(amount))
}

func QuatNlerp(q1, q2 Quat, amount float64) Quat {
	return QuatLerp(q1, q2, amount).Normalize()
}

// Performs a canonical rotation in the specified order. If the order is not
// a valid RotationOrder, this function will panic
//
// Based off the code for the Matlab function "angle2quat", though this implementation
// only supports 3 single angles as opposed to multiple angles.
func AnglesToQuat(angle1, angle2, angle3 float64, order RotationOrder) Quat {
	s := [3]float64{}
	c := [3]float64{}

	s[0], c[0] = math.Sincos(float64(angle1 / 2))
	s[1], c[1] = math.Sincos(float64(angle2 / 2))
	s[2], c[2] = math.Sincos(float64(angle3 / 2))

	ret := Quat{}
	switch order {
	case ZYX:
		ret.W = float64(c[0]*c[1]*c[2] + s[0]*s[1]*s[2])
		ret.V = Vec3{float64(c[0]*c[1]*s[2] - s[0]*s[1]*c[2]),
			float64(c[0]*s[1]*c[2] + s[0]*c[1]*s[2]),
			float64(s[0]*c[1]*c[2] - c[0]*s[1]*s[2]),
		}
	case ZYZ:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*c[1]*s[2])
		ret.V = Vec3{float64(c[0]*s[1]*s[2] - s[0]*s[1]*c[2]),
			float64(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
			float64(s[0]*c[1]*c[2] + c[0]*c[1]*s[2]),
		}
	case ZXY:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*s[1]*s[2])
		ret.V = Vec3{float64(c[0]*s[1]*c[2] - s[0]*c[1]*s[2]),
			float64(c[0]*c[1]*s[2] + s[0]*s[1]*c[2]),
			float64(c[0]*s[1]*s[2] + s[0]*c[1]*c[2]),
		}
	case ZXZ:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*c[1]*s[2])
		ret.V = Vec3{float64(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
			float64(s[0]*s[1]*c[2] - c[0]*s[1]*s[2]),
			float64(c[0]*c[1]*s[2] + s[0]*c[1]*c[2]),
		}
	case YXZ:
		ret.W = float64(c[0]*c[1]*c[2] + s[0]*s[1]*s[2])
		ret.V = Vec3{float64(c[0]*s[1]*c[2] + s[0]*c[1]*s[2]),
			float64(s[0]*c[1]*c[2] - c[0]*s[1]*s[2]),
			float64(c[0]*c[1]*s[2] - s[0]*s[1]*c[2]),
		}
	case YXY:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*c[1]*s[2])
		ret.V = Vec3{float64(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
			float64(s[0]*c[1]*c[2] + c[0]*c[1]*s[2]),
			float64(c[0]*s[1]*s[2] - s[0]*s[1]*c[2]),
		}
	case YZX:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*s[1]*s[2])
		ret.V = Vec3{float64(c[0]*c[1]*s[2] + s[0]*s[1]*c[2]),
			float64(c[0]*s[1]*s[2] + s[0]*c[1]*c[2]),
			float64(c[0]*s[1]*c[2] - s[0]*c[1]*s[2]),
		}
	case YZY:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*c[1]*s[2])
		ret.V = Vec3{float64(s[0]*s[1]*c[2] - c[0]*s[1]*s[2]),
			float64(c[0]*c[1]*s[2] + s[0]*c[1]*c[2]),
			float64(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
		}
	case XYZ:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*s[1]*s[2])
		ret.V = Vec3{float64(c[0]*s[1]*s[2] + s[0]*c[1]*c[2]),
			float64(c[0]*s[1]*c[2] - s[0]*c[1]*s[2]),
			float64(c[0]*c[1]*s[2] + s[0]*s[1]*c[2]),
		}
	case XYX:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*c[1]*s[2])
		ret.V = Vec3{float64(c[0]*c[1]*s[2] + s[0]*c[1]*c[2]),
			float64(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
			float64(s[0]*s[1]*c[2] - c[0]*s[1]*s[2]),
		}
	case XZY:
		ret.W = float64(c[0]*c[1]*c[2] + s[0]*s[1]*s[2])
		ret.V = Vec3{float64(s[0]*c[1]*c[2] - c[0]*s[1]*s[2]),
			float64(c[0]*c[1]*s[2] - s[0]*s[1]*c[2]),
			float64(c[0]*s[1]*c[2] + s[0]*c[1]*s[2]),
		}
	case XZX:
		ret.W = float64(c[0]*c[1]*c[2] - s[0]*c[1]*s[2])
		ret.V = Vec3{float64(c[0]*c[1]*s[2] + s[0]*c[1]*c[2]),
			float64(c[0]*s[1]*s[2] - s[0]*s[1]*c[2]),
			float64(c[0]*s[1]*c[2] + s[0]*s[1]*s[2]),
		}
	default:
		panic("Unsupported rotation order")
	}
	return ret
}