/*
##
#                     GNU GENERAL PUBLIC LICENSE
#                        Version 3, 29 June 2007
#
#  Copyright (C) 2007 Free Software Foundation, Inc. <http://fsf.org/>
#  Everyone is permitted to copy and distribute verbatim copies
#  of this license document, but changing it is not allowed.
#
#                             Preamble
#
#   The GNU General Public License is a free, copyleft license for
# software and other kinds of works.
#
#   The licenses for most software and other practical works are designed
# to take away your freedom to share and change the works.  By contrast,
# the GNU General Public License is intended to guarantee your freedom to
# share and change all versions of a program--to make sure it remains free
# software for all its users.  We, the Free Software Foundation, use the
# GNU General Public License for most of our software; it applies also to
# any other work released this way by its authors.  You can apply it to
# your programs, too.
#
#   When we speak of free software, we are referring to freedom, not
# price.  Our General Public Licenses are designed to make sure that you
# have the freedom to distribute copies of free software (and charge for
# them if you wish), that you receive source code or can get it if you
# want it, that you can change the software or use pieces of it in new
# free programs, and that you know you can do these things.
#
#   To protect your rights, we need to prevent others from denying you
# these rights or asking you to surrender the rights.  Therefore, you have
# certain responsibilities if you distribute copies of the software, or if
# you modify it: responsibilities to respect the freedom of others.
#
#   For example, if you distribute copies of such a program, whether
# gratis or for a fee, you must pass on to the recipients the same
# freedoms that you received.  You must make sure that they, too, receive
# or can get the source code.  And you must show them these terms so they
# know their rights.
#
#   Developers that use the GNU GPL protect your rights with two steps:
# (1) assert copyright on the software, and (2) offer you this License
# giving you legal permission to copy, distribute and/or modify it.
#
#   For the developers' and authors' protection, the GPL clearly explains
# that there is no warranty for this free software.  For both users' and
# authors' sake, the GPL requires that modified versions be marked as
# changed, so that their problems will not be attributed erroneously to
# authors of previous versions.
#
#   Some devices are designed to deny users access to install or run
# modified versions of the software inside them, although the manufacturer
# can do so.  This is fundamentally incompatible with the aim of
# protecting users' freedom to change the software.  The systematic
# pattern of such abuse occurs in the area of products for individuals to
# use, which is precisely where it is most unacceptable.  Therefore, we
# have designed this version of the GPL to prohibit the practice for those
# products.  If such problems arise substantially in other domains, we
# stand ready to extend this provision to those domains in future versions
# of the GPL, as needed to protect the freedom of users.
#
#   Finally, every program is threatened constantly by software patents.
# States should not allow patents to restrict development and use of
# software on general-purpose computers, but in those that do, we wish to
# avoid the special danger that patents applied to a free program could
# make it effectively proprietary.  To prevent this, the GPL assures that
# patents cannot be used to render the program non-free.
#
*/

package emon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"log"
	"runtime"
	"time"
	"os/exec"
	"math/rand"

)

const (
	VIEW_SOCKET = iota
	VIEW_CORE
	VIEW_THREAD
)

var (
	EDP_SCRIPT = "edp.py"

	// For emon 5.44+ readiness
	// EDP_SCRIPT = "mpp.py" 

	FILE_DMIDECODE = "dmidecode.txt"
	FILE_DISKSTAT  = "diskstat.txt"
	FILE_NETWORK   = "network.txt"
	// FILE_CHARTS    = "chart_format.txt"
	
	FILE_METRICS   = "metrics.xml"
	// for big/small uarch like alderlake
	FILE_METRICS_BIGCORE = "big.xml"
	FILE_METRICS_SMALLCORE = "small.xml"

	DEFAULT_VIEW = VIEW_CORE
	
	// DEFAULT_SUMMARY_FILE = "edpd_summary.xlsx"
	DEFAULT_SUMMARY_FILE = ""
	PYTHON = "python"

	ADDTIONAL_OPTIONS = []string{"thread-view", "uncore-view"}

)

type lock interface {
	getId() string 		// unique lock name
	IsLocked() bool		// check if locked

	Lock()				// enable lock
	Unlock()			// unlock
}

type EDP struct {
	*EmonData
	EDPPath, SummaryFile string
	PathLock             lock
	thread_number 		 uint
}

func NewEDP(p string, threads uint, emon *EmonData) *EDP {
	if p == "" {
		p, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}

	summaryFile := DEFAULT_SUMMARY_FILE
	if DEFAULT_SUMMARY_FILE == "" {
		filename := fmt.Sprintf("edp_summary_%04X.xlsx", rand.Int() & 0xFFFF)
		summaryFile = filepath.Join(emon.Path, filename)
	}else{
		summaryFile = filepath.Join(emon.Path, DEFAULT_SUMMARY_FILE)
	}
	
	lock := NewFileLock(emon.Path)
	return &EDP{emon, p, summaryFile, lock, threads}
}

func (edp *EDP) getInputFiles(edpPath, tag, value string) string {

	absPath := filepath.Join(edp.Path, value)
	// use the default inputfile
	if !fileExists(absPath) {
		absPath = filepath.Join(edpPath, value)
	}
	return fmt.Sprintf(" %s %s", tag, absPath)
}

func (edp *EDP) getMultiArchMetrics(edpPath string) string {
	
	smallCoreFile := filepath.Join(edp.Path, FILE_METRICS_SMALLCORE) 
	bigCoreFile := filepath.Join(edp.Path, FILE_METRICS_BIGCORE) 
	
	if !(fileExists(smallCoreFile) && fileExists(bigCoreFile)) {
		return edp.getInputFiles(edpPath, "-m", FILE_METRICS)
	}

	return fmt.Sprintf(" -m smallcore=%s bigcore=%s", smallCoreFile, bigCoreFile)
}

func (edp *EDP) getViewPara() string {

	views := [VIEW_THREAD + 1]string{" --socket-view",
		"--core-view", "--thread-view"}

	view := DEFAULT_VIEW
	return strings.Join(views[:view+1], " ")
}

func (edp *EDP) configableOptions() string {
	options := ""

	for _, option := range ADDTIONAL_OPTIONS {
		if (fileExists(filepath.Join(edp.Path, option))){

			options +=  fmt.Sprintf(" --%s", option)
		}
	}

	return options
}

func (edp *EDP) getCMD() string {

	edpPath := edp.EDPPath

	cmd := PYTHON + " "
	cmd += filepath.Join(edpPath, EDP_SCRIPT)

	// cmd += edp.getInputFiles(edpPath, "-f", FILE_CHARTS)
	cmd += edp.getMultiArchMetrics(edpPath)
	
	cmd += fmt.Sprintf(" -i %s", edp.EmonFile)

	if edp.EmonVFile != "" && ! edp.LatestVersion {
		cmd += fmt.Sprintf(" -j %s", edp.EmonVFile)
	}

	cmd += fmt.Sprintf(" -o %s", edp.SummaryFile)

	cmd += edp.getViewPara()
	cmd += edp.configableOptions()
	
	cmd += fmt.Sprintf(" -p %d", edp.thread_number)

	return cmd

}


func (edp *EDP) Analysis() error {
	if edp.PathLock.IsLocked() {
		return fmt.Errorf("Path %s is locked", edp.Path)
	}
	edp.PathLock.Lock()
	log.Printf("Begin to analysis %s, size, %dM", edp.Path, edp.Size()>>20)

	startTime := time.Now()
	command := edp.getCMD()

	job := exec.Command("")
	if runtime.GOOS == "windows" {
		job = exec.Command("cmd", "/C", command)
	} else {
		job = exec.Command("sh", "-c", command)
	}
	
	job.Dir = edp.Path
	err := job.Run()

	if err != nil {
		edp.PathLock.Unlock() // release emonLock for next cycle run.
		log.Printf("exit with '%s' error when runs %s", err.Error(), command)

		return fmt.Errorf("Get error when exec: %s", command)

	} else {
		stopTime := time.Now()
		log.Printf("Exported report %d seconds.", stopTime.Unix()-startTime.Unix())

		return nil
	}
}
