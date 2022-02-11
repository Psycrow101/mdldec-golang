package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Vector4 struct{ X, Y, Z, W float64 }
type Matrix3x4 [3]Vector4

type Vector4_32 struct{ X, Y, Z, W float32 }
type Matrix3x4_32 [3]Vector4_32

func (mat *Matrix3x4) From32(mat32 *Matrix3x4_32) {
	for i := 0; i < 3; i++ {
		mat[i].X = float64(mat32[i].X)
		mat[i].Y = float64(mat32[i].Y)
		mat[i].Z = float64(mat32[i].Z)
		mat[i].W = float64(mat32[i].W)
	}
}

var boneTransforms []*Matrix3x4
var worldTransform []*Matrix3x4

func matrix3x4concatTransforms(m1, m2 *Matrix3x4) *Matrix3x4 {
	var out Matrix3x4
	out[0].X = m1[0].X*m2[0].X + m1[0].Y*m2[1].X + m1[0].Z*m2[2].X
	out[0].Y = m1[0].X*m2[0].Y + m1[0].Y*m2[1].Y + m1[0].Z*m2[2].Y
	out[0].Z = m1[0].X*m2[0].Z + m1[0].Y*m2[1].Z + m1[0].Z*m2[2].Z
	out[0].W = m1[0].X*m2[0].W + m1[0].Y*m2[1].W + m1[0].Z*m2[2].W + m1[0].W
	out[1].X = m1[1].X*m2[0].X + m1[1].Y*m2[1].X + m1[1].Z*m2[2].X
	out[1].Y = m1[1].X*m2[0].Y + m1[1].Y*m2[1].Y + m1[1].Z*m2[2].Y
	out[1].Z = m1[1].X*m2[0].Z + m1[1].Y*m2[1].Z + m1[1].Z*m2[2].Z
	out[1].W = m1[1].X*m2[0].W + m1[1].Y*m2[1].W + m1[1].Z*m2[2].W + m1[1].W
	out[2].X = m1[2].X*m2[0].X + m1[2].Y*m2[1].X + m1[2].Z*m2[2].X
	out[2].Y = m1[2].X*m2[0].Y + m1[2].Y*m2[1].Y + m1[2].Z*m2[2].Y
	out[2].Z = m1[2].X*m2[0].Z + m1[2].Y*m2[1].Z + m1[2].Z*m2[2].Z
	out[2].W = m1[2].X*m2[0].W + m1[2].Y*m2[1].W + m1[2].Z*m2[2].W + m1[2].W
	return &out
}

func angleQuaternion(angles *Vector3_32) *Vector4 {
	var sr, sp, sy, cr, cp, cy float64

	sy, cy = math.Sincos(float64(angles.Z) * 0.5)
	sp, cp = math.Sincos(float64(angles.Y) * 0.5)
	sr, cr = math.Sincos(float64(angles.X) * 0.5)

	return &Vector4{
		X: sr*cp*cy - cr*sp*sy,
		Y: cr*sp*cy + sr*cp*sy,
		Z: cr*cp*sy - sr*sp*cy,
		W: cr*cp*cy + sr*sp*sy,
	}
}

func matrix3x4FromOriginQuat(quat *Vector4, origin *Vector3_32) *Matrix3x4 {
	var out Matrix3x4

	out[0].X = 1.0 - 2.0*quat.Y*quat.Y - 2.0*quat.Z*quat.Z
	out[1].X = 2.0*quat.X*quat.Y + 2.0*quat.W*quat.Z
	out[2].X = 2.0*quat.X*quat.Z - 2.0*quat.W*quat.Y

	out[0].Y = 2.0*quat.X*quat.Y - 2.0*quat.W*quat.Z
	out[1].Y = 1.0 - 2.0*quat.X*quat.X - 2.0*quat.Z*quat.Z
	out[2].Y = 2.0*quat.Y*quat.Z + 2.0*quat.W*quat.X

	out[0].Z = 2.0*quat.X*quat.Z + 2.0*quat.W*quat.Y
	out[1].Z = 2.0*quat.Y*quat.Z - 2.0*quat.W*quat.X
	out[2].Z = 1.0 - 2.0*quat.X*quat.X - 2.0*quat.Y*quat.Y

	out[0].W = float64(origin.X)
	out[1].W = float64(origin.Y)
	out[2].W = float64(origin.Z)

	return &out
}

func matrix3x4VectorTransform(m *Matrix3x4, v *Vector3_32) *Vector3_32 {
	var out [3]float64
	out[0] = float64(v.X)*m[0].X + float64(v.Y)*m[0].Y + float64(v.Z)*m[0].Z + m[0].W
	out[1] = float64(v.X)*m[1].X + float64(v.Y)*m[1].Y + float64(v.Z)*m[1].Z + m[1].W
	out[2] = float64(v.X)*m[2].X + float64(v.Y)*m[2].Y + float64(v.Z)*m[2].Z + m[2].W
	return &Vector3_32{float32(out[0]), float32(out[1]), float32(out[2])}
}

func matrix3x4VectorRotate(m *Matrix3x4, v *Vector3_32) *Vector3_32 {
	var out [3]float64
	out[0] = float64(v.X)*m[0].X + float64(v.Y)*m[0].Y + float64(v.Z)*m[0].Z
	out[1] = float64(v.X)*m[1].X + float64(v.Y)*m[1].Y + float64(v.Z)*m[1].Z
	out[2] = float64(v.X)*m[2].X + float64(v.Y)*m[2].Y + float64(v.Z)*m[2].Z
	return &Vector3_32{float32(out[0]), float32(out[1]), float32(out[2])}
}

func computeSkinMatrix(boneWeights *StudioBoneWeight) *Matrix3x4 {
	var (
		weights  [MaxBoneWeights]float64
		boneMats [MaxBoneWeights]*Matrix3x4
		bonesNum int
		total    float64
		out      = Matrix3x4{}
	)

	for _, b := range boneWeights.Bone {
		if b != -1 {
			bonesNum++
		}
	}

	for i := 0; i < bonesNum; i++ {
		boneMats[i] = worldTransform[boneWeights.Bone[i]]
		weights[i] = float64(boneWeights.Weight[i]) / 255.0
		total += weights[i]
	}

	if total < 1.0 {
		weights[0] += 1.0 - total
	}

	for i := 0; i < bonesNum; i++ {
		for j := 0; j < 3; j++ {
			out[j].X += boneMats[i][j].X * weights[i]
			out[j].Y += boneMats[i][j].Y * weights[i]
			out[j].Z += boneMats[i][j].Z * weights[i]
			out[j].W += boneMats[i][j].W * weights[i]
		}
	}

	return &out
}

func calcBonePosition(anim *Anim, bone *StudioBone, frame int) [6]float64 {
	var (
		motion   [6]float64
		animVals []*AnimValue
		value    float64
		j        int
	)

	for i := 0; i < 6; i++ {
		motion[i] = float64(bone.Value[i])
		animVals = anim.AnimValues[i]
		if animVals == nil {
			continue
		}

		j = frame
		for _, av := range animVals {
			if j >= int(av.Total) {
				j -= int(av.Total)
				continue
			}
			if int(av.Valid) > j {
				value = float64(av.Values[j])
			} else {
				value = float64(av.Values[av.Valid-1])
			}
			break
		}
		motion[i] += value * float64(bone.Scale[i])
	}

	return motion
}

func properBoneRotationZ(seq *Sequence, motion *[6]float64, frame int, angle float64) {
	motion[0] += float64(frame) / float64(seq.FramesNum) * float64(seq.LinerMovement.X)
	motion[1] += float64(frame) / float64(seq.FramesNum) * float64(seq.LinerMovement.Y)
	motion[2] += float64(frame) / float64(seq.FramesNum) * float64(seq.LinerMovement.Z)

	rot := angle * math.Pi / 180.0
	s, c := math.Sin(rot), math.Cos(rot)
	x, y := motion[0], motion[1]
	motion[0] = c*x - s*y
	motion[1] = s*x + c*y
	motion[5] += rot
}

func clipRotations(val *float64) {
	for *val >= math.Pi {
		*val -= math.Pi * 2.0
	}
	for *val < -math.Pi {
		*val += math.Pi * 2.0
	}
}

func writeNodes(writer *bufio.Writer, bones []*StudioBone) {
	writer.WriteString("nodes\n")
	for i, b := range bones {
		writer.WriteString(fmt.Sprintf("%3d \"%s\" %d\n", i, b.Name, b.Parent))
	}
	writer.WriteString("end\n")
}

func writeSkeleton(writer *bufio.Writer, bones []*StudioBone) {
	writer.WriteString("skeleton\n")
	writer.WriteString("time 0\n")
	for i, b := range bones {
		writer.WriteString(fmt.Sprintf("%3d", i))
		for _, v := range b.Value {
			writer.WriteString(fmt.Sprintf(" %f", v))
		}
		writer.WriteString("\n")
	}
	writer.WriteString("end\n")
}

func writeTriangleInfo(writer *bufio.Writer, model *Model, mdl *Mdl,
	skinRef uint32, triangle [3]*StudioTriangle, isEvenStrip bool) {

	var (
		indices              [3]int
		vertIndex, normIndex uint16
		boneIndex            byte
		vert                 *StudioTriangle
		u, v                 float32
		vertPos, vertNorm    *Vector3_32
		vertWeightsNum       int
		vertWeight           *StudioBoneWeight
	)

	if isEvenStrip {
		indices[0] = 1
		indices[1] = 2
		indices[2] = 0
	} else {
		indices[0] = 0
		indices[1] = 1
		indices[2] = 2
	}

	texture := mdl.Textures[skinRef]
	s := 1.0 / float64(texture.Width)
	t := 1.0 / float64(texture.Height)

	writer.WriteString(fmt.Sprintf("%s\n", texture.Name))

	for i := 0; i < 3; i++ {
		vert = triangle[indices[i]]
		vertIndex = vert.VertexIndex
		normIndex = vert.NormalIndex
		boneIndex = model.VerticesInfo[vertIndex]

		u = float32((float64(vert.S)) * s)
		v = float32(1.0 - float64(vert.T)*t)

		if mdl.Header.Flags&StudioHasBoneWeights != 0 {
			vertWeight = &model.VerticesWeights[vertIndex]
			mat := computeSkinMatrix(vertWeight)
			vertPos = matrix3x4VectorTransform(mat, &model.Vertices[vertIndex])
			vertNorm = matrix3x4VectorRotate(mat, &model.Normals[normIndex])
			vertNorm.Normalize()

			writer.WriteString(fmt.Sprintf("%3d %f %f %f %f %f %f %f %f",
				boneIndex,
				vertPos.X, vertPos.Y, vertPos.Z,
				vertNorm.X, vertNorm.Y, vertNorm.Z,
				u, v))

			vertWeightsNum = 0

			for _, b := range vertWeight.Bone {
				if b != -1 {
					vertWeightsNum++
				}
			}

			if vertWeightsNum > 0 {
				writer.WriteString(fmt.Sprintf(" %d", vertWeightsNum))
				for b := 0; b < vertWeightsNum; b++ {
					writer.WriteString(fmt.Sprintf(" %d %f",
						vertWeight.Bone[b], float32(vertWeight.Weight[b])/255.0))
				}
			}
			writer.WriteString("\n")

		} else {
			vertPos = matrix3x4VectorTransform(boneTransforms[boneIndex], &model.Vertices[vertIndex])
			vertNorm = matrix3x4VectorRotate(boneTransforms[boneIndex], &model.Normals[normIndex])
			vertNorm.Normalize()

			writer.WriteString(fmt.Sprintf("%3d %f %f %f %f %f %f %f %f\n",
				boneIndex,
				vertPos.X, vertPos.Y, vertPos.Z,
				vertNorm.X, vertNorm.Y, vertNorm.Z,
				u, v))
		}
	}
}

func writeTriangles(writer *bufio.Writer, model *Model, mdl *Mdl) {
	var triangle [3]*StudioTriangle

	writer.WriteString("triangles\n")
	for _, me := range model.Meshes {
		skinRef := me.SkinRef
		for _, tri := range me.Triangles {
			if tri.IsStrip {
				for i, v := range tri.Vertices {
					switch {
					case i == 0:
						triangle[0] = v
					case i == 1:
						triangle[2] = v
					case i == 2:
						triangle[1] = v
						writeTriangleInfo(writer, model, mdl, skinRef, triangle, false)
					default:
						triangle[2], triangle[1] = triangle[1], v
						writeTriangleInfo(writer, model, mdl, skinRef, triangle, false)
					}
				}
			} else {
				for i, v := range tri.Vertices {
					switch {
					case i == 0:
						triangle[0] = v
					case i == 1:
						triangle[2] = v
					case i == 2:
						triangle[1] = v
						writeTriangleInfo(writer, model, mdl, skinRef, triangle, true)
					case i%2 > 0:
						triangle[0], triangle[2] = triangle[2], v
						writeTriangleInfo(writer, model, mdl, skinRef, triangle, false)
					default:
						triangle[0], triangle[1] = triangle[1], v
						writeTriangleInfo(writer, model, mdl, skinRef, triangle, true)
					}
				}
			}
		}
	}
	writer.WriteString("end\n")
}

func writeFrameInfo(writer *bufio.Writer, seq *Sequence, bones []*StudioBone, blendId int, frame int) {
	writer.WriteString(fmt.Sprintf("time %d\n", frame))

	for i, bone := range bones {
		motion := calcBonePosition(seq.Anims[blendId*len(bones)+i], bone, frame)

		if bone.Parent == -1 {
			properBoneRotationZ(seq, &motion, frame, 270.0)
		}
		clipRotations(&motion[3])
		clipRotations(&motion[4])
		clipRotations(&motion[5])

		writer.WriteString(fmt.Sprintf("%3d  ", i))
		for j := 0; j < 6; j++ {
			writer.WriteString(fmt.Sprintf(" %f", motion[j]))
		}

		writer.WriteString("\n")
	}
}

func writeAnimations(writer *bufio.Writer, bones []*StudioBone, seq *Sequence, blendId int) {
	writer.WriteString("skeleton\n")

	for i := 0; i < int(seq.FramesNum); i++ {
		writeFrameInfo(writer, seq, bones, blendId, i)
	}

	writer.WriteString("end\n")
}

func saveReferences(outPath string, mdl *Mdl) error {
	var (
		err               error
		filePath, smdName string
		file              *os.File
		writer            *bufio.Writer
	)

	boneTransforms = make([]*Matrix3x4, mdl.Header.BonesNum)

	for i, bone := range mdl.Bones {
		quat := angleQuaternion(&Vector3_32{bone.Value[3], bone.Value[4], bone.Value[5]})
		boneTransforms[i] = matrix3x4FromOriginQuat(quat,
			&Vector3_32{bone.Value[0], bone.Value[1], bone.Value[2]})

		if bone.Parent > -1 {
			boneTransforms[i] = matrix3x4concatTransforms(boneTransforms[bone.Parent],
				boneTransforms[i])
		}
	}

	if mdl.Header.Flags&StudioHasBoneInfo != 0 {
		worldTransform = make([]*Matrix3x4, mdl.Header.BonesNum)
		poseToBone := new(Matrix3x4)

		for i, boneInfo := range mdl.BonesInfo {
			poseToBone.From32(&boneInfo.PoseToBone)
			worldTransform[i] = matrix3x4concatTransforms(boneTransforms[i], poseToBone)
		}
	}

	for _, bp := range mdl.BodyParts {
		for _, m := range bp.Models {
			if m.Name.String() == "blank" {
				continue
			}

			func() {
				smdName = strings.TrimSuffix(m.Name.String(), ".smd") + ".smd"
				filePath = filepath.Join(outPath, smdName)

				if err = os.RemoveAll(filePath); err != nil {
					printError(err)
					return
				}

				file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					printError(err)
					return
				}
				defer file.Close()

				writer = bufio.NewWriter(file)
				defer writer.Flush()

				writer.WriteString("version 1\n")

				writeNodes(writer, mdl.Bones)
				writeSkeleton(writer, mdl.Bones)
				writeTriangles(writer, m, mdl)

				fmt.Printf("Reference: %s\n", smdName)
			}()
		}
	}
	return nil
}

func saveSequences(outPath string, mdl *Mdl) error {
	var (
		err               error
		filePath, smdName string
		file              *os.File
		writer            *bufio.Writer
	)

	for _, seq := range mdl.Sequences {
		for i := 0; i < int(seq.BlendsNum); i++ {
			func() {
				smdName = strings.TrimSuffix(seq.Label.String(), ".smd")
				if seq.BlendsNum > 1 {
					smdName = fmt.Sprintf("%s_blend%d", smdName, i+1)
				}
				smdName += ".smd"
				filePath = filepath.Join(outPath, smdName)

				if err = os.RemoveAll(filePath); err != nil {
					printError(err)
					return
				}

				file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					printError(err)
					return
				}
				defer file.Close()

				writer = bufio.NewWriter(file)
				defer writer.Flush()

				writer.WriteString("version 1\n")

				writeNodes(writer, mdl.Bones)
				writeAnimations(writer, mdl.Bones, seq, i)

				fmt.Printf("Sequence: %s\n", smdName)
			}()
		}
	}
	return nil
}

func saveSMDs(destPath string, mdl *Mdl) error {
	if err := saveReferences(destPath, mdl); err != nil {
		return err
	}

	sequencesPath := filepath.Join(destPath, "anims")
	if err := createDirectory(sequencesPath); err != nil {
		printError(err)
		return err
	}

	if err := saveSequences(sequencesPath, mdl); err != nil {
		return err
	}
	return nil
}
