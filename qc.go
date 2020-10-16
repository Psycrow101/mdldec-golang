package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var activityNames = []string{
	"ACT_RESET",
	"ACT_IDLE",
	"ACT_GUARD",
	"ACT_WALK",
	"ACT_RUN",
	"ACT_FLY",
	"ACT_SWIM",
	"ACT_HOP",
	"ACT_LEAP",
	"ACT_FALL",
	"ACT_LAND",
	"ACT_STRAFE_LEFT",
	"ACT_STRAFE_RIGHT",
	"ACT_ROLL_LEFT",
	"ACT_ROLL_RIGHT",
	"ACT_TURN_LEFT",
	"ACT_TURN_RIGHT",
	"ACT_CROUCH",
	"ACT_CROUCHIDLE",
	"ACT_STAND",
	"ACT_USE",
	"ACT_SIGNAL1",
	"ACT_SIGNAL2",
	"ACT_SIGNAL3",
	"ACT_TWITCH",
	"ACT_COWER",
	"ACT_SMALL_FLINCH",
	"ACT_BIG_FLINCH",
	"ACT_RANGE_ATTACK1",
	"ACT_RANGE_ATTACK2",
	"ACT_MELEE_ATTACK1",
	"ACT_MELEE_ATTACK2",
	"ACT_RELOAD",
	"ACT_ARM",
	"ACT_DISARM",
	"ACT_EAT",
	"ACT_DIESIMPLE",
	"ACT_DIEBACKWARD",
	"ACT_DIEFORWARD",
	"ACT_DIEVIOLENT",
	"ACT_BARNACLE_HIT",
	"ACT_BARNACLE_PULL",
	"ACT_BARNACLE_CHOMP",
	"ACT_BARNACLE_CHEW",
	"ACT_SLEEP",
	"ACT_INSPECT_FLOOR",
	"ACT_INSPECT_WALL",
	"ACT_IDLE_ANGRY",
	"ACT_WALK_HURT",
	"ACT_RUN_HURT",
	"ACT_HOVER",
	"ACT_GLIDE",
	"ACT_FLY_LEFT",
	"ACT_FLY_RIGHT",
	"ACT_DETECT_SCENT",
	"ACT_SNIFF",
	"ACT_BITE",
	"ACT_THREAT_DISPLAY",
	"ACT_FEAR_DISPLAY",
	"ACT_EXCITED",
	"ACT_SPECIAL_ATTACK1",
	"ACT_SPECIAL_ATTACK2",
	"ACT_COMBAT_IDLE",
	"ACT_WALK_SCARED",
	"ACT_RUN_SCARED",
	"ACT_VICTORY_DANCE",
	"ACT_DIE_HEADSHOT",
	"ACT_DIE_CHESTSHOT",
	"ACT_DIE_GUTSHOT",
	"ACT_DIE_BACKSHOT",
	"ACT_FLINCH_HEAD",
	"ACT_FLINCH_CHEST",
	"ACT_FLINCH_STOMACH",
	"ACT_FLINCH_LEFTARM",
	"ACT_FLINCH_RIGHTARM",
	"ACT_FLINCH_LEFTLEG",
	"ACT_FLINCH_RIGHTLEG",
	"ACT_VM_NONE",
	"ACT_VM_DEPLOY",
	"ACT_VM_DEPLOY_EMPTY",
	"ACT_VM_HOLSTER",
	"ACT_VM_HOLSTER_EMPTY",
	"ACT_VM_IDLE1",
	"ACT_VM_IDLE2",
	"ACT_VM_IDLE3",
	"ACT_VM_RANGE_ATTACK1",
	"ACT_VM_RANGE_ATTACK2",
	"ACT_VM_RANGE_ATTACK3",
	"ACT_VM_MELEE_ATTACK1",
	"ACT_VM_MELEE_ATTACK2",
	"ACT_VM_MELEE_ATTACK3",
	"ACT_VM_SHOOT_EMPTY",
	"ACT_VM_START_RELOAD",
	"ACT_VM_RELOAD",
	"ACT_VM_RELOAD_EMPTY",
	"ACT_VM_TURNON",
	"ACT_VM_TURNOFF",
	"ACT_VM_PUMP",
	"ACT_VM_PUMP_EMPTY",
	"ACT_VM_START_CHARGE",
	"ACT_VM_CHARGE",
	"ACT_VM_OVERLOAD",
	"ACT_VM_IDLE_EMPTY",
}

func getMotionTypeString(motionType int, isComposite bool) string {
	if isComposite {
		var sb strings.Builder
		if motionType&StudioMotionX > 0 {
			sb.WriteString(" X")
		}
		if motionType&StudioMotionY > 0 {
			sb.WriteString(" Y")
		}
		if motionType&StudioMotionZ > 0 {
			sb.WriteString(" Z")
		}
		if motionType&StudioMotionXR > 0 {
			sb.WriteString(" XR")
		}
		if motionType&StudioMotionYR > 0 {
			sb.WriteString(" YR")
		}
		if motionType&StudioMotionZR > 0 {
			sb.WriteString(" ZR")
		}
		if motionType&StudioMotionLX > 0 {
			sb.WriteString(" LX")
		}
		if motionType&StudioMotionLY > 0 {
			sb.WriteString(" LY")
		}
		if motionType&StudioMotionLZ > 0 {
			sb.WriteString(" LZ")
		}
		if motionType&StudioMotionAX > 0 {
			sb.WriteString(" AX")
		}
		if motionType&StudioMotionAY > 0 {
			sb.WriteString(" AY")
		}
		if motionType&StudioMotionAZ > 0 {
			sb.WriteString(" AZ")
		}
		if motionType&StudioMotionAXR > 0 {
			sb.WriteString(" AXR")
		}
		if motionType&StudioMotionAYR > 0 {
			sb.WriteString(" AYR")
		}
		if motionType&StudioMotionAZR > 0 {
			sb.WriteString(" AZR")
		}
		return sb.String()
	} else {
		motionType &= StudioMotionTypes
		switch motionType {
		case StudioMotionX:
			return "X"
		case StudioMotionY:
			return "Y"
		case StudioMotionZ:
			return "Z"
		case StudioMotionXR:
			return "XR"
		case StudioMotionYR:
			return "YR"
		case StudioMotionZR:
			return "ZR"
		case StudioMotionLX:
			return "LX"
		case StudioMotionLY:
			return "LY"
		case StudioMotionLZ:
			return "LZ"
		case StudioMotionAX:
			return "AX"
		case StudioMotionAY:
			return "AY"
		case StudioMotionAZ:
			return "AZ"
		case StudioMotionAXR:
			return "AXR"
		case StudioMotionAYR:
			return "AYR"
		case StudioMotionAZR:
			return "AZR"
		}
	}
	return ""
}

func writeBodyGroupInfo(writer *bufio.Writer, mdl *Mdl) {
	writer.WriteString("\n// reference mesh(es)\n")

	for _, bg := range mdl.BodyParts {
		bodyGroupName := strings.TrimSuffix(bg.Name.String(), ".smd")

		if bg.ModelsNum == 1 {
			modelName := strings.TrimSuffix(bg.Models[0].Name.String(), ".smd")
			writer.WriteString(fmt.Sprintf("$body \"%s\" \"%s\"\n\n", bodyGroupName, modelName))
			continue
		}

		writer.WriteString(fmt.Sprintf("$bodygroup \"%s\"\n{\n", bodyGroupName))
		for _, m := range bg.Models {
			modelName := strings.TrimSuffix(m.Name.String(), ".smd")
			if modelName == "blank" {
				writer.WriteString("\tblank\n")
				continue
			}
			writer.WriteString(fmt.Sprintf("\tstudio \"%s\"\n", modelName))
		}

		writer.WriteString("}\n\n")
	}
}

func writeTextureRenderMode(writer *bufio.Writer, mdl *Mdl) {
	for _, tex := range mdl.Textures {
		if tex.Flags&StudioNfFlatshade > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"flatshade\" \n", tex.Name))
		}
		if tex.Flags&StudioNfChrome > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"chrome\" \n", tex.Name))
		}
		if tex.Flags&StudioNfFullbright > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"fullbright\" \n", tex.Name))
		}
		if tex.Flags&StudioNfNomips > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"nomips\" \n", tex.Name))
		}
		if tex.Flags&StudioNfNosmooth > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"alpha\" \n", tex.Name))
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"nosmooth\" \n", tex.Name))
		}
		if tex.Flags&StudioNfAdditive > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"additive\" \n", tex.Name))
		}
		if tex.Flags&StudioNfMasked > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"masked\" \n", tex.Name))
		}
		if tex.Flags&(StudioNfMasked|StudioNfSolid) > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"masked_solid\" \n", tex.Name))
		}
		if tex.Flags&StudioNfTwoside > 0 {
			writer.WriteString(fmt.Sprintf("$texrendermode \"%s\" \"twoside\" \n", tex.Name))
		}
	}
}

func writeSkinFamilyInfo(writer *bufio.Writer, mdl *Mdl) {
	if mdl.Header.SkinFamiliesNum < 2 {
		return
	}

	writer.WriteString(fmt.Sprintf("\n// %d skin families\n", mdl.Header.SkinFamiliesNum))
	writer.WriteString("$texturegroup skinfamilies \n{\n")

	for _, sf := range *mdl.Skins {
		writer.WriteString("\t{")
		for i, sr := range sf {
			for _, sfm := range *mdl.Skins {
				if sr != sfm[i] {
					writer.WriteString(fmt.Sprintf(" \"%s\" ", mdl.Textures[sr].Name))
					break
				}
			}
		}
		writer.WriteString("}\n")
	}
	writer.WriteString("}\n")
}

func writeAttachmentInfo(writer *bufio.Writer, mdl *Mdl) {
	if mdl.Header.AttachmentsNum == 0 {
		return
	}

	writer.WriteString(fmt.Sprintf("\n// %d attachment(s)\n", mdl.Header.AttachmentsNum))

	for i, a := range mdl.Attachments {
		bone := mdl.Bones[a.Bone]
		writer.WriteString(fmt.Sprintf("$attachment %d \"%s\" %f %f %f\n",
			i, bone.Name, a.Origins.X, a.Origins.Y, a.Origins.Z))
	}
}

func writeControllerInfo(writer *bufio.Writer, mdl *Mdl) {
	if mdl.Header.BoneControllersNum == 0 {
		return
	}

	writer.WriteString(fmt.Sprintf("\n// %d bone controller(s)\n", mdl.Header.BoneControllersNum))

	for _, bc := range mdl.BoneControllers {
		bone := mdl.Bones[bc.Bone]
		motionType := getMotionTypeString(int(bc.Type) & ^StudioMotionRLoop, false)
		writer.WriteString(fmt.Sprintf("$controller %d \"%s\" %s %f %f\n",
			bc.Index, bone.Name, motionType, bc.Start, bc.End))
	}
}

func writeHitBoxInfo(writer *bufio.Writer, mdl *Mdl) {
	if mdl.Header.HitBoxesNum == 0 {
		return
	}

	writer.WriteString(fmt.Sprintf("\n// %d hit box(es)\n", mdl.Header.HitBoxesNum))

	for _, hb := range mdl.HitBoxes {
		bone := mdl.Bones[hb.Bone]
		writer.WriteString(fmt.Sprintf("$hbox %d \"%s\" %f %f %f %f %f %f\n",
			hb.Group, bone.Name,
			hb.BBMin.X, hb.BBMin.Y, hb.BBMin.Z,
			hb.BBMax.X, hb.BBMax.Y, hb.BBMax.Z))
	}
}

func writeSequenceInfo(writer *bufio.Writer, mdl *Mdl) {
	if mdl.Header.SequenceGroupsNum > 1 {
		writer.WriteString("\n$sequencegroupsize 64\n")
	}

	if mdl.Header.SequencesNum > 0 {
		writer.WriteString(fmt.Sprintf("\n// %d animation sequence(s)\n", mdl.Header.SequencesNum))
	}

	for _, seq := range mdl.Sequences {
		writer.WriteString(fmt.Sprintf("$sequence \"%s\" ", seq.Label))

		if seq.BlendsNum > 1 {
			if seq.BlendsNum > 2 {
				writer.WriteString("{\n")
				for j := 1; j <= int(seq.BlendsNum); j++ {
					writer.WriteString("          ")
					writer.WriteString(fmt.Sprintf("\"%s_blend%d\" ", seq.Label, j))
					writer.WriteString("\n")
				}
				writer.WriteString("          ")
			} else {
				writer.WriteString(fmt.Sprintf("\"%s_blend1\" ", seq.Label))
				writer.WriteString(fmt.Sprintf("\"%s_blend2\" ", seq.Label))
			}
			writer.WriteString(fmt.Sprintf("blend %s %.0f %.0f",
				getMotionTypeString(int(seq.BlendTypes[0]), false),
				seq.BlendStart[0], seq.BlendEnd[0]))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\"", seq.Label))
		}

		if seq.MotionType > 0 {
			writer.WriteString(getMotionTypeString(int(seq.MotionType), true))
		}

		writer.WriteString(fmt.Sprintf(" fps %.0f ", seq.FPS))

		if seq.Flags == 1 {
			writer.WriteString("loop ")
		}

		if seq.Activity > 0 {
			if int(seq.Activity) < len(activityNames) {
				writer.WriteString(fmt.Sprintf("%s %d ",
					activityNames[seq.Activity], seq.ActWight))
			} else {
				fmt.Printf("WARNING: Sequence %s has a custom activity flag (ACT_%d %d).\n",
					seq.Label, seq.Activity, seq.ActWight)
				writer.WriteString(fmt.Sprintf("ACT_%d %d ",
					seq.Activity, seq.ActWight))
			}
		}

		if seq.EntryNode != 0 && seq.ExitNode != 0 {
			if seq.EntryNode == seq.ExitNode {
				writer.WriteString(fmt.Sprintf("node %d ", seq.EntryNode))
			} else if seq.NodeFlags != 0 {
				writer.WriteString(fmt.Sprintf("rtransition %d %d ",
					seq.EntryNode, seq.ExitNode))
			} else {
				writer.WriteString(fmt.Sprintf("transition %d %d ",
					seq.EntryNode, seq.ExitNode))
			}
		}

		if seq.EventsNum > 2 {
			writer.WriteString("{\n ")
			for _, ev := range seq.Events {
				if seq.BlendsNum <= 2 {
					writer.WriteString(" ")
				} else {
					writer.WriteString("          ")
				}

				writer.WriteString(fmt.Sprintf("{ event %d %d", ev.Event, ev.Frame))
				if ev.Options[0] != 0 {
					writer.WriteString(fmt.Sprintf(" \"%s\"", ev.Options))
				}
				writer.WriteString(" }\n ")
			}
			writer.WriteString("}")
		} else {
			for _, ev := range seq.Events {
				writer.WriteString(fmt.Sprintf("{ event %d %d", ev.Event, ev.Frame))
				if ev.Options[0] != 0 {
					writer.WriteString(fmt.Sprintf(" \"%s\"", ev.Options))
				}
				writer.WriteString(" } ")
			}
		}

		writer.WriteString("\n")

		if seq.BlendsNum > 2 {
			writer.WriteString("}\n")
		}

		if seq.PivotsNum > 0 {
			fmt.Printf("WARNING: Sequence %s uses %d foot pivots, feature not supported.\n",
				seq.Label, seq.PivotsNum)
		}
	}
}

func saveQCScript(outPath string, mdl *Mdl) error {
	var (
		err    error
		file   *os.File
		writer *bufio.Writer
	)

	if err = os.RemoveAll(outPath); err != nil {
		return err
	}

	file, err = os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer = bufio.NewWriter(file)
	defer writer.Flush()

	writer.WriteString("/*\n")
	writer.WriteString("==============================================================================\n\n")
	writer.WriteString(fmt.Sprintf("QC script generated by Half-Life Studio Model Decompiler on Go %s\n\n",
		appVersion))
	writer.WriteString(fmt.Sprintf("%s\n\n", mdl.FilePath))
	writer.WriteString("Original internal name:\n")
	writer.WriteString(fmt.Sprintf("\"%s\"\n\n", mdl.Header.Name))
	writer.WriteString("==============================================================================\n")
	writer.WriteString("*/\n\n")

	writer.WriteString(fmt.Sprintf("$modelname \"%s\"\n", filepath.Base(mdl.FilePath)))
	writer.WriteString("$cd \".\\\"\n")
	writer.WriteString("$cdtexture \".\\\"\n")
	writer.WriteString("$scale 1.0\n")
	writer.WriteString("$cliptotextures\n\n")

	if mdl.Header.TexturesNum == 0 {
		writer.WriteString("$externaltextures\n")
	}

	if mdl.Header.Flags != 0 {
		writer.WriteString(fmt.Sprintf("$flags %d\n", mdl.Header.Flags))
		fmt.Printf("WARNING: This model uses the $flags keyword set to %d\n", mdl.Header.Flags)
	}

	writer.WriteString("\n")
	writer.WriteString(fmt.Sprintf("$bbox %f %f %f",
		mdl.Header.BBMin.X, mdl.Header.BBMin.Y, mdl.Header.BBMin.Z))
	writer.WriteString(fmt.Sprintf(" %f %f %f\n",
		mdl.Header.BBMax.X, mdl.Header.BBMax.Y, mdl.Header.BBMax.Z))
	writer.WriteString(fmt.Sprintf("$cbox %f %f %f",
		mdl.Header.BBMin.X, mdl.Header.BBMin.Y, mdl.Header.BBMin.Z))
	writer.WriteString(fmt.Sprintf(" %f %f %f\n",
		mdl.Header.BBMax.X, mdl.Header.BBMax.Y, mdl.Header.BBMax.Z))
	writer.WriteString(fmt.Sprintf("$eyeposition %f %f %f\n",
		mdl.Header.EyePosition.X, mdl.Header.EyePosition.Y, mdl.Header.EyePosition.Z))
	writer.WriteString("\n")

	writeBodyGroupInfo(writer, mdl)
	writeTextureRenderMode(writer, mdl)
	writeSkinFamilyInfo(writer, mdl)
	writeAttachmentInfo(writer, mdl)
	writeControllerInfo(writer, mdl)
	writeHitBoxInfo(writer, mdl)
	writeSequenceInfo(writer, mdl)

	writer.WriteString("\n// End of QC script.\n")

	fmt.Printf("QC Script: %s\n", outPath)
	return nil
}
