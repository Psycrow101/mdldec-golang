package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func isValidName(name string) bool {
	return checkReg(name, `^[\w\-.][\w\-. ]*$`)
}

func checkReg(str, pattern string) bool {
	if matched, err := regexp.Match(pattern, []byte(str)); matched && err == nil {
		return true
	}
	return false
}

func fixNames(mdl *Mdl) {
	var str string

	texPattern := `(_texture)(\d+)(.bmp)`
	modelPattern := `(_body)(\d+)(_)(\d+)`
	seqPattern := `(_seq)(\d+)`

	var usedNames [3]map[string]bool
	for i := 0; i < len(usedNames); i++ {
		usedNames[i] = make(map[string]bool)
	}
	isNameUsed := func(name string, t int) bool {
		if _, ok := usedNames[t][name]; ok {
			return true
		}
		return false
	}

	for i, tex := range mdl.Textures {
		str = tex.Name.String()
		if !isValidName(str) || checkReg(str, texPattern) || isNameUsed(str, 0) {
			str = fmt.Sprintf("_texture%d.bmp", i+1)
			tex.Name.FromString(str)
			usedNames[0][str] = true
		}
	}

	for i, bp := range mdl.BodyParts {
		str = bp.Name.String()
		if !isValidName(str) {
			str = fmt.Sprintf("_bodypart%d", i+1)
			bp.Name.FromString(str)
		}
		for j, m := range bp.Models {
			str = m.Name.String()
			if !isValidName(str) || checkReg(str, modelPattern) || isNameUsed(str, 1) {
				str = fmt.Sprintf("_body%d_%d", i+1, j+1)
				m.Name.FromString(str)
				usedNames[1][str] = true
			}
		}
	}

	for i, seq := range mdl.Sequences {
		str = seq.Label.String()
		if !isValidName(str) || checkReg(str, seqPattern) || isNameUsed(str, 2) {
			str = fmt.Sprintf("_seq%d", i+1)
			seq.Label.FromString(str)
			usedNames[2][str] = true
		}
	}
}

func loadSeqMDL(modelPath string, mdl *Mdl, seqGroupId uint32) error {
	var (
		err  error
		file *os.File
	)

	file, err = os.Open(modelPath)
	if err != nil {
		return err
	}
	defer file.Close()

	studioHdr := new(StudioHdr)
	err = binary.Read(file, binary.LittleEndian, studioHdr)
	if err != nil {
		return err
	}

	if studioHdr.Ident != SeqIdent {
		return errors.New(fmt.Sprintf("%s is not a valid sequence file", modelPath))
	}

	for _, seq := range mdl.Sequences {
		if seq.SeqGroup != seqGroupId || seq.Anims != nil {
			continue
		}
		seq.SeqGroup = 0
		if err = seq.readAnims(file, mdl.Header.BonesNum); err != nil {
			return err
		}
	}

	return nil
}

func loadMDL(modelPath string) (*Mdl, error) {
	var (
		err  error
		file *os.File
	)

	ext := filepath.Ext(modelPath)
	if len(ext) == 0 {
		return nil, errors.New("source file does not have extension")
	}

	if ext != ".mdl" {
		return nil, errors.New("only .mdl-files is supported")
	}

	file, err = os.Open(modelPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	studioHdr := new(StudioHdr)
	err = binary.Read(file, binary.LittleEndian, studioHdr)
	if err != nil {
		return nil, err
	}

	if studioHdr.Ident != MdlIdent {
		if studioHdr.Ident == SeqIdent {
			return nil, errors.New(fmt.Sprintf("%s is not a main HL model file", modelPath))
		} else {
			return nil, errors.New(fmt.Sprintf("%s is not a valid HL model file", modelPath))
		}
	}

	if studioHdr.Version != StudioVersion {
		return nil, errors.New(fmt.Sprintf("%s has unknown Studio MDL format version", modelPath))
	}

	mdl := new(Mdl)
	mdl.Header = studioHdr
	mdl.FilePath = modelPath

	if err = mdl.ReadBones(file); err != nil {
		return nil, err
	}
	if err = mdl.ReadBoneControllers(file); err != nil {
		return nil, err
	}
	if err = mdl.ReadHitBoxes(file); err != nil {
		return nil, err
	}
	if err = mdl.ReadSequences(file); err != nil {
		return nil, err
	}
	if err = mdl.ReadBodyParts(file); err != nil {
		return nil, err
	}
	if err = mdl.ReadAttachments(file); err != nil {
		return nil, err
	}
	if err = mdl.ReadTextures(file); err != nil {
		return nil, err
	}
	if err = mdl.ReadSkins(file); err != nil {
		return nil, err
	}

	if studioHdr.TexturesNum == 0 {
		mdlTPath := strings.TrimSuffix(modelPath, ".mdl") + "T.mdl"
		var mdlT *Mdl
		mdlT, err = loadMDL(mdlTPath)
		if err != nil {
			return nil, err
		} else {
			mdl.Textures = mdlT.Textures
			mdl.Skins = mdlT.Skins
		}
	}

	if studioHdr.SequenceGroupsNum > 1 {
		for i := 1; i < int(studioHdr.SequenceGroupsNum); i++ {
			seqPath := strings.TrimSuffix(modelPath, ".mdl")
			seqPath += fmt.Sprintf("%02d.mdl", i)
			err = loadSeqMDL(seqPath, mdl, uint32(i))
			if err != nil {
				return nil, err
			}
		}
	}

	if studioHdr.HitBoxesNum > MaxHitboxes {
		fmt.Printf("[WARNING] Invalid hitboxes number (%d) \n", studioHdr.HitBoxesNum)
		studioHdr.HitBoxesNum = 0
	} else if studioHdr.HitBoxesOff+studioHdr.HitBoxesNum*68 > studioHdr.Length {
		fmt.Printf("[WARNING] Invalid hitboxes offset (%d) \n", studioHdr.HitBoxesOff)
		studioHdr.HitBoxesNum = 0
	}

	fixNames(mdl)

	return mdl, nil
}

func showHelp(appName string) {
	fmt.Printf("usage: %s source_file\n", appName)
	fmt.Printf("       %s source_file target_directory\n", appName)
}

func main() {
	fmt.Printf("\nHalf-Life Studio Model Decompiler %s on Go\n", appVersion)
	fmt.Println("--------------------------------------------------")
	defer fmt.Println("--------------------------------------------------")

	args := os.Args
	argsNum := len(args)
	var destPath string

	if argsNum == 1 {
		showHelp(args[0])
		return
	} else if argsNum == 2 {
		destPath = filepath.Join(filepath.Dir(args[1]), "decomp_"+filepath.Base(args[1]))
	} else {
		destPath = args[2]
	}

	if err := createDirectory(destPath); err != nil {
		printError(err)
		return
	}

	if mdl, err := loadMDL(args[1]); err != nil {
		printError(err)
	} else {
		wg := &sync.WaitGroup{}
		wg.Add(3)

		go func() {
			defer wg.Done()
			qcFileName := filepath.Base(args[1])
			qcFileName = qcFileName[:len(qcFileName)-3] + "qc"
			if err = saveQCScript(filepath.Join(destPath, qcFileName), mdl); err != nil {
				printError(err)
			}
		}()

		go func() {
			defer wg.Done()
			if err = saveSMDs(destPath, mdl); err != nil {
				printError(err)
			}
		}()

		go func() {
			defer wg.Done()
			texturesPath := filepath.Join(destPath, "textures")
			if err := createDirectory(texturesPath); err != nil {
				printError(err)
				return
			}
			
			if err = saveTextures(texturesPath, mdl); err != nil {
				printError(err)
			}
		}()

		wg.Wait()
	}

	fmt.Println("Done.")
}
