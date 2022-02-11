package main

import (
	"encoding/binary"
	"io"
	"math"
	"os"
)

const MdlIdent = 0x54534449
const SeqIdent = 0x51534449
const StudioVersion = 10

const MaxHitboxes = 512
const MaxBoneWeights = 4

// client-side model flags
const (
	StudioHasBoneInfo    = 1 << 30
	StudioHasBoneWeights = 1 << 31
)

// lighting & rendermode options
const (
	StudioNfFlatshade = 1 << iota
	StudioNfChrome
	StudioNfFullbright
	StudioNfNomips    // ignore mip-maps
	StudioNfNosmooth  // don't smooth tangent space
	StudioNfAdditive  // rendering with additive mode
	StudioNfMasked    // use texture with alpha channel
	StudioNfNormalmap // indexed normalmap
	StudioNfSolid     = 1 << (iota + 3)
	StudioNfTwoside             // render mesh as twosided
	StudioNfColormap  = 1 << 30 // internal system flag
	StudioNfUvCoords  = 1 << 31 // using half-float coords instead of ST
)

// motion flags
const (
	StudioMotionX = 1 << iota
	StudioMotionY
	StudioMotionZ
	StudioMotionXR
	StudioMotionYR
	StudioMotionZR
	StudioMotionLX
	StudioMotionLY
	StudioMotionLZ
	StudioMotionAX
	StudioMotionAY
	StudioMotionAZ
	StudioMotionAXR
	StudioMotionAYR
	StudioMotionAZR
	StudioMotionTypes = 0x7FFF
	StudioMotionRLoop = 0x8000 // controller that wraps shortest distance
)

type Vector3_32 struct{ X, Y, Z float32 }

type Bytes32 [32]byte
type Bytes64 [64]byte

func (v *Vector3_32) Normalize() {
	vecLen := float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
	if vecLen > 0.0 {
		vecLen = 1.0 / vecLen
	}
	v.X *= vecLen
	v.Y *= vecLen
	v.Z *= vecLen
}

func (bytes Bytes64) String() string {
	return bytesToString(bytes[:])
}

func (bytes *Bytes64) FromString(str string) {
	strLen := len(str)
	if strLen > 63 {
		strLen = 63
	}
	for i := 0; i < strLen; i++ {
		bytes[i] = str[i]
	}
	bytes[strLen] = 0
}

func (bytes Bytes32) String() string {
	return bytesToString(bytes[:])
}

func (bytes *Bytes32) FromString(str string) {
	strLen := len(str)
	if strLen > 31 {
		strLen = 31
	}
	for i := 0; i < strLen; i++ {
		bytes[i] = str[i]
	}
	bytes[len(str)] = 0
}

func bytesToString(bytes []byte) string {
	var n int
	for i, b := range bytes {
		if b == 0 {
			break
		}
		n = i
	}
	return string(bytes[:n+1])
}

type StudioHdr struct {
	Ident   uint32
	Version uint32

	Name   Bytes64
	Length uint32

	EyePosition Vector3_32 // ideal eye position
	Min         Vector3_32 // ideal movement hull size
	Max         Vector3_32

	BBMin Vector3_32 // clipping bounding box
	BBMax Vector3_32

	Flags uint32

	BonesNum    uint32 // bones
	BonesOffset uint32

	BoneControllersNum uint32 // bone controllers
	BoneControllersOff uint32

	HitBoxesNum uint32 // complex bounding boxes
	HitBoxesOff uint32 // complex bounding boxes

	SequencesNum uint32 // animation sequences
	SequencesOff uint32

	SequenceGroupsNum uint32 // demand loaded sequences
	SequenceGroupsOff uint32

	TexturesNum     uint32 // raw textures
	TexturesOff     uint32
	TexturesDataOff uint32

	SkinRefsNum     uint32 // replaceable textures
	SkinFamiliesNum uint32
	SkinsOff        uint32

	BodyPartsNum uint32
	BodyPartsOff uint32

	AttachmentsNum uint32 // queryable attachable points
	AttachmentsOff uint32

	StudioHdr2Off uint32
	SoundsOff     uint32

	SoundGroupsNum uint32
	SoundGroupsOff uint32

	TransitionsNum uint32
	TransitionsOff uint32
}

type StudioBoneController struct {
	Bone  int32  // -1 == 0
	Type  uint32 // X, Y, Z, XR, YR, ZR, M
	Start float32
	End   float32
	Rest  uint32 // byte index value at rest
	Index uint32 // 0-3 user set controller, 4 mouth
}

type StudioBone struct {
	Name            Bytes32 // bone name for symbolic links
	Parent          int32   // parent bone index
	Flags           uint32
	BoneControllers [6]uint32  // bone controller index, -1 == none
	Value           [6]float32 // default DoF values
	Scale           [6]float32 // scale for delta DoF values
}

type StudioHitBox struct {
	Bone  uint32
	Group uint32     // intersection group
	BBMin Vector3_32 // bounding box
	BBMax Vector3_32
}

type StudioSequence struct {
	Label Bytes32 // sequence label
	FPS   float32 // frames per second
	Flags uint32  // looping/non-looping flags

	Activity uint32
	ActWight int32

	EventsNum uint32
	EventsOff uint32

	FramesNum uint32

	PivotsNum uint32 // number of foot pivots
	PivotsOff uint32

	MotionType       uint32
	MotionBone       uint32
	LinerMovement    Vector3_32
	AutoMovePosOff   uint32
	AutoMoveAngleOff uint32

	BBMin Vector3_32 // per sequence bounding box
	BBMax Vector3_32

	BlendsNum uint32
	AnimOff   uint32 // StudioAnimation pointer relative to start of sequence group data
	// [blend][bone][X, Y, Z, XR, YR, ZR]

	BlendTypes  [2]uint32  // X, Y, Z, XR, YR, ZR
	BlendStart  [2]float32 // starting value
	BlendEnd    [2]float32
	BlendParent int32

	SeqGroup uint32 // sequence group for demand loading

	EntryNode int32  // transition node at entry
	ExitNode  int32  // transition node at exit
	NodeFlags uint32 // transition rules

	NextSeq int32 // auto advancing sequences
}

type StudioAnim struct {
	Offsets [6]uint16
}

type StudioEvent struct {
	Frame   uint32
	Event   int32
	Type    uint32
	Options Bytes64
}

type StudioTexture struct {
	Name   Bytes64
	Flags  uint32
	Width  uint32
	Height uint32
	Offset uint32
}

type StudioBodyPart struct {
	Name      Bytes64
	ModelsNum uint32
	Base      uint32
	ModelsOff uint32 // index into models array
}

type StudioModel struct {
	Name Bytes64
	Type int32

	BoundingRadius float32

	MeshesNum uint32
	MeshesOff uint32

	VertsNum       uint32 // number of unique vertices
	VertsInfoOff   uint32 // vertex bone info
	VertsOff       uint32 // vertex vector3s
	NormalsNum     uint32 // number of unique surface normals
	NormalsInfoOff uint32 // normal bone info
	NormalsOff     uint32 // normal vector3s

	BlendVertInfoOff uint32 // boneweighted vertex info
	BlendNormInfoOff uint32 // boneweighted normal info
}

type StudioMesh struct {
	TrianglesNum uint32
	TrianglesOff uint32
	SkinRef      uint32
	NormalsNum   uint32
	NormalsOff   uint32
}

type StudioBoneWeight struct {
	Weight [MaxBoneWeights]uint8
	Bone   [MaxBoneWeights]int8
}

type StudioBoneInfo struct {
	PoseToBone Matrix3x4_32
	QAlignment Vector4_32
	ProcType   int32
	ProcIndex  int32
	Quat       Vector4_32
	Reserved   [10]int32
}

type StudioTriangle struct {
	VertexIndex uint16 // index into vertex array
	NormalIndex uint16 // index into normal array
	S, T        int16  // s, t position on skin
}

type StudioAttachment struct {
	Name    Bytes32
	Type    uint32
	Bone    uint32
	Origins Vector3_32
	Vectors [3]Vector3_32
}

type Sequence struct {
	StudioSequence
	Events []*StudioEvent
	Anims  []*Anim
}

type Anim struct {
	AnimValues [6][]*AnimValue
}

type AnimValue struct {
	Valid  uint8
	Total  uint8
	Values []int16
}

type Texture struct {
	StudioTexture
	Indices []byte
	Pallets [256 * 3]byte
}

type BodyPart struct {
	StudioBodyPart
	Models []*Model
}

type Model struct {
	StudioModel
	Meshes          []*Mesh
	Vertices        []Vector3_32
	VerticesInfo    []byte
	Normals         []Vector3_32
	VerticesWeights []StudioBoneWeight
}

type Mesh struct {
	StudioMesh
	Triangles []*Triangle
}

type Triangle struct {
	IsStrip  bool
	Vertices []*StudioTriangle
}

type Mdl struct {
	FilePath        string
	Header          *StudioHdr
	Bones           []*StudioBone
	BonesInfo       []*StudioBoneInfo
	BoneControllers []*StudioBoneController
	HitBoxes        []*StudioHitBox
	Sequences       []*Sequence
	Textures        []*Texture
	Skins           *[][]uint16
	BodyParts       []*BodyPart
	Attachments     []*StudioAttachment
}

func (mdl *Mdl) ReadBones(file *os.File) error {
	if _, err := file.Seek(int64(mdl.Header.BonesOffset), 0); err != nil {
		return err
	}
	var bones = make([]*StudioBone, mdl.Header.BonesNum)
	for i := 0; i < int(mdl.Header.BonesNum); i++ {
		b := new(StudioBone)
		if err := binary.Read(file, binary.LittleEndian, b); err != nil {
			return err
		}
		bones[i] = b
	}
	mdl.Bones = bones

	if mdl.Header.Flags&StudioHasBoneInfo != 0 {
		var bonesInfo = make([]*StudioBoneInfo, mdl.Header.BonesNum)
		for i := 0; i < int(mdl.Header.BonesNum); i++ {
			bi := new(StudioBoneInfo)
			if err := binary.Read(file, binary.LittleEndian, bi); err != nil {
				return err
			}
			bonesInfo[i] = bi
		}
		mdl.BonesInfo = bonesInfo
	}

	return nil
}

func (mdl *Mdl) ReadBoneControllers(file *os.File) error {
	if _, err := file.Seek(int64(mdl.Header.BoneControllersOff), 0); err != nil {
		return err
	}
	var boneControllers = make([]*StudioBoneController, mdl.Header.BoneControllersNum)
	for i := 0; i < int(mdl.Header.BoneControllersNum); i++ {
		bc := new(StudioBoneController)
		if err := binary.Read(file, binary.LittleEndian, bc); err != nil {
			return err
		}
		boneControllers[i] = bc
	}
	mdl.BoneControllers = boneControllers
	return nil
}

func (mdl *Mdl) ReadHitBoxes(file *os.File) error {
	if _, err := file.Seek(int64(mdl.Header.HitBoxesOff), 0); err != nil {
		return err
	}
	var hitBoxes = make([]*StudioHitBox, mdl.Header.HitBoxesNum)
	for i := 0; i < int(mdl.Header.HitBoxesNum); i++ {
		hb := new(StudioHitBox)
		if err := binary.Read(file, binary.LittleEndian, hb); err != nil {
			return err
		}
		hitBoxes[i] = hb
	}
	mdl.HitBoxes = hitBoxes
	return nil
}

func (mdl *Mdl) ReadSequences(file *os.File) error {
	if _, err := file.Seek(int64(mdl.Header.SequencesOff), 0); err != nil {
		return err
	}
	var sequences = make([]*Sequence, mdl.Header.SequencesNum)
	for i := 0; i < int(mdl.Header.SequencesNum); i++ {
		seq := new(Sequence)
		if err := binary.Read(file, binary.LittleEndian, &seq.StudioSequence); err != nil {
			return err
		}
		curFileOff, _ := file.Seek(0, io.SeekCurrent)
		if err := seq.readEvents(file); err != nil {
			return err
		}
		if err := seq.readAnims(file, mdl.Header.BonesNum); err != nil {
			return err
		}
		file.Seek(curFileOff, 0)
		sequences[i] = seq
	}
	mdl.Sequences = sequences
	return nil
}

func (seq *Sequence) readEvents(file *os.File) error {
	if _, err := file.Seek(int64(seq.EventsOff), 0); err != nil {
		return err
	}
	var events = make([]*StudioEvent, seq.EventsNum)
	for i := 0; i < int(seq.EventsNum); i++ {
		ev := new(StudioEvent)
		if err := binary.Read(file, binary.LittleEndian, ev); err != nil {
			return err
		}
		events[i] = ev
	}
	seq.Events = events
	return nil
}

func (seq *Sequence) readAnims(file *os.File, bonesNum uint32) error {
	if seq.SeqGroup > 0 {
		return nil
	}

	animOff := int64(seq.AnimOff)

	animsNum := seq.BlendsNum * bonesNum
	if _, err := file.Seek(animOff, 0); err != nil {
		return err
	}
	var anims = make([]*Anim, animsNum)

	var studioAnims = make([]StudioAnim, animsNum)
	for i := 0; i < int(animsNum); i++ {
		a := StudioAnim{}
		if err := binary.Read(file, binary.LittleEndian, &a); err != nil {
			return err
		}
		studioAnims[i] = a
	}

	for i, a := range studioAnims {
		anim := new(Anim)
		for j := 0; j < 6; j++ {
			if a.Offsets[j] == 0 {
				continue
			}
			animValues := make([]*AnimValue, 0, seq.FramesNum)
			file.Seek(animOff+int64(i*12)+int64(a.Offsets[j]), 0)

			f := cap(animValues)
			for f > 0 {
				av := new(AnimValue)
				binary.Read(file, binary.LittleEndian, &av.Valid)
				binary.Read(file, binary.LittleEndian, &av.Total)
				av.Values = make([]int16, av.Valid)
				binary.Read(file, binary.LittleEndian, &av.Values)
				animValues = append(animValues, av)
				f -= int(av.Total)
			}
			anim.AnimValues[j] = animValues
		}
		anims[i] = anim
	}

	seq.Anims = anims
	return nil
}

func (mdl *Mdl) ReadTextures(file *os.File) error {
	if _, err := file.Seek(int64(mdl.Header.TexturesOff), 0); err != nil {
		return err
	}
	var textures = make([]*Texture, mdl.Header.TexturesNum)
	for i := 0; i < int(mdl.Header.TexturesNum); i++ {
		t := new(Texture)
		if err := binary.Read(file, binary.LittleEndian, &t.StudioTexture); err != nil {
			return err
		}
		curFileOff, _ := file.Seek(0, io.SeekCurrent)
		t.Indices = make([]byte, t.Width*t.Height)
		if _, err := file.Seek(int64(t.Offset), 0); err != nil {
			return err
		}
		if err := binary.Read(file, binary.LittleEndian, &t.Indices); err != nil {
			return err
		}
		if err := binary.Read(file, binary.LittleEndian, &t.Pallets); err != nil {
			return err
		}
		file.Seek(curFileOff, 0)
		textures[i] = t
	}
	mdl.Textures = textures
	return nil
}

func (mdl *Mdl) ReadSkins(file *os.File) error {
	if _, err := file.Seek(int64(mdl.Header.SkinsOff), 0); err != nil {
		return err
	}
	var skins = make([][]uint16, mdl.Header.SkinFamiliesNum)

	for i := 0; i < int(mdl.Header.SkinFamiliesNum); i++ {
		skins[i] = make([]uint16, mdl.Header.SkinRefsNum)
		if err := binary.Read(file, binary.LittleEndian, &skins[i]); err != nil {
			return err
		}
	}
	mdl.Skins = &skins
	return nil
}

func (mdl *Mdl) ReadBodyParts(file *os.File) error {
	if _, err := file.Seek(int64(mdl.Header.BodyPartsOff), 0); err != nil {
		return err
	}
	var bodyParts = make([]*BodyPart, mdl.Header.BodyPartsNum)
	for i := 0; i < int(mdl.Header.BodyPartsNum); i++ {
		bp := new(BodyPart)
		if err := binary.Read(file, binary.LittleEndian, &bp.StudioBodyPart); err != nil {
			return err
		}
		curFileOff, _ := file.Seek(0, io.SeekCurrent)
		if err := bp.readModels(file, mdl.Header.Flags&StudioHasBoneWeights != 0); err != nil {
			return err
		}
		file.Seek(curFileOff, 0)
		bodyParts[i] = bp
	}
	mdl.BodyParts = bodyParts
	return nil
}

func (bp *BodyPart) readModels(file *os.File, hasBoneWeights bool) error {
	if _, err := file.Seek(int64(bp.ModelsOff), 0); err != nil {
		return err
	}
	var models = make([]*Model, bp.ModelsNum)
	for i := 0; i < int(bp.ModelsNum); i++ {
		m := new(Model)

		if err := binary.Read(file, binary.LittleEndian, &m.StudioModel); err != nil {
			return err
		}
		curFileOff, _ := file.Seek(0, io.SeekCurrent)
		if err := m.readMeshes(file); err != nil {
			return err
		}

		if _, err := file.Seek(int64(m.VertsOff), 0); err != nil {
			return err
		}
		m.Vertices = make([]Vector3_32, m.VertsNum)
		if err := binary.Read(file, binary.LittleEndian, &m.Vertices); err != nil {
			return err
		}

		if _, err := file.Seek(int64(m.VertsInfoOff), 0); err != nil {
			return err
		}
		m.VerticesInfo = make([]byte, m.VertsNum)
		if err := binary.Read(file, binary.LittleEndian, &m.VerticesInfo); err != nil {
			return err
		}

		if _, err := file.Seek(int64(m.NormalsOff), 0); err != nil {
			return err
		}
		m.Normals = make([]Vector3_32, m.NormalsNum)
		if err := binary.Read(file, binary.LittleEndian, &m.Normals); err != nil {
			return err
		}

		if hasBoneWeights {
			if _, err := file.Seek(int64(m.BlendVertInfoOff), 0); err != nil {
				return err
			}
			m.VerticesWeights = make([]StudioBoneWeight, m.VertsNum)
			if err := binary.Read(file, binary.LittleEndian, &m.VerticesWeights); err != nil {
				return err
			}
		}

		file.Seek(curFileOff, 0)
		models[i] = m
	}
	bp.Models = models
	return nil
}

func (m *Model) readMeshes(file *os.File) error {
	if _, err := file.Seek(int64(m.MeshesOff), 0); err != nil {
		return err
	}
	var meshes = make([]*Mesh, m.MeshesNum)
	for i := 0; i < int(m.MeshesNum); i++ {
		me := new(Mesh)
		if err := binary.Read(file, binary.LittleEndian, &me.StudioMesh); err != nil {
			return err
		}
		curFileOff, _ := file.Seek(0, io.SeekCurrent)
		if err := me.readTriangles(file); err != nil {
			return err
		}
		file.Seek(curFileOff, 0)
		meshes[i] = me
	}
	m.Meshes = meshes
	return nil
}

func (mesh *Mesh) readTriangles(file *os.File) error {
	var trianglesNum int16

	if _, err := file.Seek(int64(mesh.TrianglesOff), 0); err != nil {
		return err
	}

	var triangles = make([]*Triangle, 0, mesh.TrianglesNum)
	for {
		if err := binary.Read(file, binary.LittleEndian, &trianglesNum); err != nil {
			return err
		}
		if trianglesNum == 0 {
			break
		}

		tri := new(Triangle)
		if trianglesNum < 0 {
			trianglesNum = -trianglesNum
			tri.IsStrip = true
		}
		tri.Vertices = make([]*StudioTriangle, trianglesNum)

		for i := 0; i < int(trianglesNum); i++ {
			v := new(StudioTriangle)
			if err := binary.Read(file, binary.LittleEndian, v); err != nil {
				return err
			}
			tri.Vertices[i] = v
		}

		triangles = append(triangles, tri)
	}

	mesh.Triangles = triangles
	return nil
}

func (mdl *Mdl) ReadAttachments(file *os.File) error {
	if _, err := file.Seek(int64(mdl.Header.AttachmentsOff), 0); err != nil {
		return err
	}
	var attachments = make([]*StudioAttachment, mdl.Header.AttachmentsNum)
	for i := 0; i < int(mdl.Header.AttachmentsNum); i++ {
		a := new(StudioAttachment)
		if err := binary.Read(file, binary.LittleEndian, a); err != nil {
			return err
		}
		attachments[i] = a
	}
	mdl.Attachments = attachments
	return nil
}
