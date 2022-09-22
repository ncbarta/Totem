package main

import (
	"Totem/vscoservice"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	httpClient   = http.DefaultClient
	home, _      = os.UserHomeDir()
	desktop      = home + "/Desktop"
	totemPath    = desktop + "/Totem"
	trackingFile = totemPath + "/Tracking.yaml"
)

var (
	rootCMD = &cobra.Command{
		Use:   "totem",
		Short: "Multi-user OSINT Tool",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
)

func main() {
	// Create /Totem if it does not exist
	if _, err := os.Stat(totemPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(totemPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	os.Chdir(totemPath)
	trackingProfiles := deserializeTrackingProfiles()

	runCMD := &cobra.Command{
		Use:   "run",
		Short: "Runs all active profiles, or runs profiles selected.",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				runAllActiveTrackingProfiles(trackingProfiles)
			} else {
				for _, a := range args {
					runTrackingProfile(a, trackingProfiles)
				}
			}
		},
	}

	printCMD := &cobra.Command{
		Use:   "print",
		Short: "Prints out all profiles",
		Run: func(cmd *cobra.Command, args []string) {
			printTrackingProfiles(trackingProfiles)
		},
	}

	bioCMD := &cobra.Command{
		Use:   "bio",
		Short: "Prints out a profile's bios",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			getBioInfoForTrackingProfile(args[0], trackingProfiles)
		},
	}

	rootCMD.AddCommand(runCMD)
	rootCMD.AddCommand(printCMD)
	rootCMD.AddCommand(bioCMD)

	Execute()

	serializeTrackingProfiles(&trackingProfiles)
}

func Execute() {
	if err := rootCMD.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runAllActiveTrackingProfiles(trackingProfiles []vscoservice.VSCOTrackingProfile) {
	for i := range trackingProfiles {
		if trackingProfiles[i].Active {
			getUser(&trackingProfiles[i])
		}
	}
}

func runTrackingProfile(targetName string, trackingProfiles []vscoservice.VSCOTrackingProfile) {
	for i := range trackingProfiles {
		if trackingProfiles[i].TargetName == targetName {
			getUser(&trackingProfiles[i])
		}
	}
}

func getBioInfoForTrackingProfile(targetName string, trackingProfiles []vscoservice.VSCOTrackingProfile) {
	// For each account, print out the bio history
	for i := range trackingProfiles {
		if trackingProfiles[i].TargetName == targetName {
			userpath := totemPath + "/" + trackingProfiles[i].TargetName
			for k := range trackingProfiles[i].Accounts {
				accountPath := userpath + "/" + trackingProfiles[i].Accounts[k].Username
				os.Chdir(accountPath)

				service := vscoservice.New(&trackingProfiles[i].Accounts[k], userpath)
				service.PrintBio(false)

				os.Chdir("..")
			}
		}
	}
}

func printTrackingProfiles(trackingProfiles []vscoservice.VSCOTrackingProfile) {
	for _, tp := range trackingProfiles {
		fmt.Println("--------" + tp.TargetName + "--------")
		fmt.Println("Active:", tp.Active)
		fmt.Println("Accounts:")
		for _, a := range tp.Accounts {
			fmt.Println(" +", a.Username, "("+strconv.Itoa(a.UserID)+")")
		}
	}
}

func deserializeTrackingProfiles() []vscoservice.VSCOTrackingProfile {
	var trackingProfiles []vscoservice.VSCOTrackingProfile

	if _, err := os.Stat(trackingFile); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	bytes, err := os.ReadFile(trackingFile)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(bytes, &trackingProfiles)
	if err != nil {
		panic(err)
	}

	return trackingProfiles
}

func serializeTrackingProfiles(profiles *[]vscoservice.VSCOTrackingProfile) {
	bytes, err := yaml.Marshal(profiles)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(trackingFile, bytes, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

// Runs vscoservice on a TrackingProfile & all of it's accounts
func getUser(tp *vscoservice.VSCOTrackingProfile) {
	userpath := totemPath + "/" + tp.TargetName

	if _, err := os.Stat(userpath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(userpath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	os.Chdir(userpath)

	for i, a := range tp.Accounts {
		accountPath := userpath + "/" + a.Username
		if _, err := os.Stat(accountPath); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(accountPath, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		os.Chdir(accountPath)

		service := vscoservice.New(&tp.Accounts[i], userpath)
		service.CheckBio()
		service.CheckProfileImage()
		service.CheckGalleryMedia()
		service.CheckCollectionMedia()

		os.Chdir("..")
	}
}
