package grsync

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Rsync wraps one rsync command.
type Rsync struct {
	// source is the transfer source
	Source string
	// destination is the transfer destination
	Destination string

	cmd *exec.Cmd
}

// RsyncOptions configures an rsync command.
type RsyncOptions struct {
	// rsyncBinaryPath overrides the rsync executable path
	RsyncBinaryPath string
	// rsyncPath sets the command used to invoke rsync remotely
	RsyncPath string
	// verbose increases command output
	Verbose bool
	// quiet suppresses non-error messages
	Quiet bool
	// checksum compares file checksums instead of modification time and size
	Checksum bool
	// archive enables archive mode without preserving hard links, ACLs, or xattrs
	Archive bool
	// recursive descends into directories
	Recursive bool
	// relative preserves relative path names
	Relative bool
	// noImpliedDirs omits implied directories when relative mode is enabled
	NoImpliedDirs bool
	// update skips receiver files that are newer
	Update bool
	// inplace updates destination files in place
	Inplace bool
	// append adds data to shorter destination files
	Append bool
	// appendVerify verifies existing data before appending
	AppendVerify bool
	// dirs transfers directories without recursing
	Dirs bool
	// links copies symbolic links as symbolic links
	Links bool
	// copyLinks copies each symbolic-link target
	CopyLinks bool
	// copyUnsafeLinks copies targets of symbolic links that escape the tree
	CopyUnsafeLinks bool
	// safeLinks ignores symbolic links that escape the tree
	SafeLinks bool
	// copyDirLinks copies symbolic links to directories as directories
	CopyDirLinks bool
	// keepDirLinks treats receiver symbolic links to directories as directories
	KeepDirLinks bool
	// hardLinks preserves hard links
	HardLinks bool
	// perms preserves permissions
	Perms bool
	// noPerms disables permission preservation
	NoPerms bool
	// executability preserves executability
	Executability bool
	// chmod changes received file and directory permissions
	CHMOD os.FileMode
	// access-control lists are preserved when enabled
	ACLs bool
	// xattrs preserves extended attributes
	XAttrs bool
	// owner preserves file ownership
	Owner bool
	// noOwner disables owner preservation
	NoOwner bool
	// group preserves group ownership
	Group bool
	// noGroup disables group preservation
	NoGroup bool
	// devices preserves device files
	Devices bool
	// specials preserves special files
	Specials bool
	// times preserves modification times
	Times bool
	// noTimes disables modification-time preservation
	NoTimes bool
	// omitDirTimes omits directories from modification-time preservation
	OmitDirTimes bool
	// super enables receiver super-user activities
	Super bool
	// fakeSuper stores privileged attributes in extended attributes
	FakeSuper bool
	// sparse handles sparse files efficiently
	Sparse bool
	// dryRun reports changes without applying them
	DryRun bool
	// wholeFile disables the delta-transfer algorithm
	WholeFile bool
	// oneFileSystem prevents crossing filesystem boundaries
	OneFileSystem bool
	// blockSize sets the checksum block size
	BlockSize int
	// rsh sets the remote shell command
	Rsh string
	// existing skips creation of receiver files
	Existing bool
	// ignoreExisting skips receiver files that already exist
	IgnoreExisting bool
	// removeSourceFiles removes synchronized source files
	RemoveSourceFiles bool
	// delete removes extraneous destination files
	Delete bool
	// deleteBefore removes destination files before transfer
	DeleteBefore bool
	// deleteDuring removes destination files during transfer
	DeleteDuring bool
	// deleteDelay defers discovered deletions until after transfer
	DeleteDelay bool
	// deleteAfter discovers and removes destination files after transfer
	DeleteAfter bool
	// deleteExcluded also removes excluded destination files
	DeleteExcluded bool
	// ignoreErrors allows deletion despite input or output errors
	IgnoreErrors bool
	// force allows deletion of non-empty directories
	Force bool
	// maxDelete limits the number of deleted files
	MaxDelete int
	// maxSize excludes files larger than this size
	MaxSize int
	// minSize excludes files smaller than this size
	MinSize int
	// partial preserves partially transferred files
	Partial bool
	// partialDir stores partial transfers in this directory
	PartialDir string
	// delayUpdates moves updated files into place after transfer
	DelayUpdates bool
	// pruneEmptyDirs removes empty directory chains from the file list
	PruneEmptyDirs bool
	// numericIDs preserves numeric user and group IDs
	NumericIDs bool
	// timeout sets the input and output timeout in seconds
	Timeout int
	// contimeout sets the daemon connection timeout in seconds
	Contimeout int
	// ignoreTimes transfers files even when size and modification time match
	IgnoreTimes bool
	// sizeOnly compares files by size
	SizeOnly bool
	// modifyWindow selects reduced-accuracy modification-time comparison
	ModifyWindow bool
	// tempDir selects the directory for temporary files
	TempDir string
	// fuzzy searches for a similar receiver file as a transfer basis
	Fuzzy bool
	// compareDest compares receiver files relative to this directory
	CompareDest string
	// copyDest copies unchanged files from this directory
	CopyDest string
	// linkDest hard-links unchanged files from this directory
	LinkDest string
	// compress compresses transferred file data
	Compress bool
	// compressLevel sets the compression level
	CompressLevel int
	// skipCompress excludes matching suffixes from compression
	SkipCompress []string
	// version-control exclusions enable the standard CVS rules
	CVSExclude bool
	// stats emits transfer statistics
	Stats bool
	// humanReadable formats numbers for people
	HumanReadable bool
	// progress emits transfer progress
	Progress bool
	// passwordFile provides the daemon-access password
	PasswordFile string
	// bandwidthLimit limits socket input and output bandwidth
	BandwidthLimit int
	// info selects informational messages
	Info string
	// exclude lists excluded remote paths
	Exclude []string
	// include lists included remote paths
	Include []string
	// filter sets one rsync filter rule
	Filter string
	// chown sets received ownership
	Chown string

	// ipv4 restricts sockets to IPv4
	IPv4 bool
	// ipv6 restricts sockets to IPv6
	IPv6 bool

	// outFormat enables the configured output format
	OutFormat bool
}

// StdoutPipe returns a pipe that will be connected to the command's
// standard output when the command starts.
func (r Rsync) StdoutPipe() (io.ReadCloser, error) {
	return r.cmd.StdoutPipe()
}

// StderrPipe returns a pipe that will be connected to the command's
// standard error when the command starts.
func (r Rsync) StderrPipe() (io.ReadCloser, error) {
	return r.cmd.StderrPipe()
}

// Run starts the rsync command and waits for it to finish.
func (r Rsync) Run() error {
	if !strings.Contains(r.Destination, ":") && !isExist(r.Destination) {
		if err := createDir(r.Destination); err != nil {
			return err
		}
	}
	if err := r.cmd.Start(); err != nil {
		return err
	}
	return r.cmd.Wait()
}

// NewRsync constructs an rsync command.
func NewRsync(
	ctx context.Context,
	source string,
	destination string,
	options RsyncOptions,
) *Rsync {
	arguments := append(getArguments(options), source, destination)
	binaryPath := "rsync"
	if options.RsyncBinaryPath != "" {
		binaryPath = options.RsyncBinaryPath
	}
	return &Rsync{
		Source:      source,
		Destination: destination,
		cmd:         exec.CommandContext(ctx, binaryPath, arguments...),
	}
}

func getArguments(options RsyncOptions) []string {
	arguments := []string{}

	if options.RsyncPath != "" {
		arguments = append(arguments, "--rsync-path", options.RsyncPath)
	}

	if options.Verbose {
		arguments = append(arguments, "--verbose")
	}

	if options.Checksum {
		arguments = append(arguments, "--checksum")
	}

	if options.Quiet {
		arguments = append(arguments, "--quiet")
	}

	if options.Archive {
		arguments = append(arguments, "--archive")
	}

	if options.Recursive {
		arguments = append(arguments, "--recursive")
	}

	if options.Relative {
		arguments = append(arguments, "--relative")
	}

	if options.NoImpliedDirs {
		arguments = append(arguments, "--no-implied-dirs")
	}

	if options.Update {
		arguments = append(arguments, "--update")
	}

	if options.Inplace {
		arguments = append(arguments, "--inplace")
	}

	if options.Append {
		arguments = append(arguments, "--append")
	}

	if options.AppendVerify {
		arguments = append(arguments, "--append-verify")
	}

	if options.Dirs {
		arguments = append(arguments, "--dirs")
	}

	if options.Links {
		arguments = append(arguments, "--links")
	}

	if options.CopyLinks {
		arguments = append(arguments, "--copy-links")
	}

	if options.CopyUnsafeLinks {
		arguments = append(arguments, "--copy-unsafe-links")
	}

	if options.SafeLinks {
		arguments = append(arguments, "--safe-links")
	}

	if options.CopyDirLinks {
		arguments = append(arguments, "--copy-dir-links")
	}

	if options.KeepDirLinks {
		arguments = append(arguments, "--keep-dir-links")
	}

	if options.HardLinks {
		arguments = append(arguments, "--hard-links")
	}

	if options.Perms {
		arguments = append(arguments, "--perms")
	}

	if options.NoPerms {
		arguments = append(arguments, "--no-perms")
	}

	if options.Executability {
		arguments = append(arguments, "--executability")
	}

	if options.ACLs {
		arguments = append(arguments, "--acls")
	}

	if options.XAttrs {
		arguments = append(arguments, "--xattrs")
	}

	if options.Owner {
		arguments = append(arguments, "--owner")
	}

	if options.NoOwner {
		arguments = append(arguments, "--no-owner")
	}

	if options.Group {
		arguments = append(arguments, "--group")
	}

	if options.NoGroup {
		arguments = append(arguments, "--no-group")
	}

	if options.Devices {
		arguments = append(arguments, "--devices")
	}

	if options.Specials {
		arguments = append(arguments, "--specials")
	}

	if options.Times {
		arguments = append(arguments, "--times")
	}

	if options.NoTimes {
		arguments = append(arguments, "--no-times")
	}

	if options.OmitDirTimes {
		arguments = append(arguments, "--omit-dir-times")
	}

	if options.Super {
		arguments = append(arguments, "--super")
	}

	if options.FakeSuper {
		arguments = append(arguments, "--fake-super")
	}

	if options.Sparse {
		arguments = append(arguments, "--sparse")
	}

	if options.DryRun {
		arguments = append(arguments, "--dry-run")
	}

	if options.WholeFile {
		arguments = append(arguments, "--whole-file")
	}

	if options.OneFileSystem {
		arguments = append(arguments, "--one-file-system")
	}

	if options.BlockSize > 0 {
		arguments = append(arguments, "--block-size", strconv.Itoa(options.BlockSize))
	}

	if options.Rsh != "" {
		arguments = append(arguments, "--rsh", options.Rsh)
	}

	if options.Existing {
		arguments = append(arguments, "--existing")
	}

	if options.IgnoreExisting {
		arguments = append(arguments, "--ignore-existing")
	}

	if options.RemoveSourceFiles {
		arguments = append(arguments, "--remove-source-files")
	}

	if options.Delete {
		arguments = append(arguments, "--delete")
	}

	if options.DeleteBefore {
		arguments = append(arguments, "--delete-before")
	}

	if options.DeleteDuring {
		arguments = append(arguments, "--delete-during")
	}

	if options.DeleteDelay {
		arguments = append(arguments, "--delete-delay")
	}

	if options.DeleteAfter {
		arguments = append(arguments, "--delete-after")
	}

	if options.DeleteExcluded {
		arguments = append(arguments, "--delete-excluded")
	}

	if options.IgnoreErrors {
		arguments = append(arguments, "--ignore-errors")
	}

	if options.Force {
		arguments = append(arguments, "--force")
	}

	if options.MaxDelete > 0 {
		arguments = append(arguments, "--max-delete", strconv.Itoa(options.MaxDelete))
	}

	if options.MaxSize > 0 {
		arguments = append(arguments, "--max-size", strconv.Itoa(options.MaxSize))
	}

	if options.MinSize > 0 {
		arguments = append(arguments, "--min-size", strconv.Itoa(options.MinSize))
	}

	if options.Partial {
		arguments = append(arguments, "--partial")
	}

	if options.PartialDir != "" {
		arguments = append(arguments, "--partial-dir", options.PartialDir)
	}

	if options.DelayUpdates {
		arguments = append(arguments, "--delay-updates")
	}

	if options.PruneEmptyDirs {
		arguments = append(arguments, "--prune-empty-dirs")
	}

	if options.NumericIDs {
		arguments = append(arguments, "--numeric-ids")
	}

	if options.Timeout > 0 {
		arguments = append(arguments, "--timeout", strconv.Itoa(options.Timeout))
	}

	if options.Contimeout > 0 {
		arguments = append(arguments, "--contimeout", strconv.Itoa(options.Contimeout))
	}

	if options.IgnoreTimes {
		arguments = append(arguments, "--ignore-times")
	}

	if options.SizeOnly {
		arguments = append(arguments, "--size-only")
	}

	if options.ModifyWindow {
		arguments = append(arguments, "--modify-window")
	}

	if options.TempDir != "" {
		arguments = append(arguments, "--temp-dir", options.TempDir)
	}

	if options.Fuzzy {
		arguments = append(arguments, "--fuzzy")
	}

	if options.CompareDest != "" {
		arguments = append(arguments, "--compare-dest", options.CompareDest)
	}

	if options.CopyDest != "" {
		arguments = append(arguments, "--copy-dest", options.CopyDest)
	}

	if options.LinkDest != "" {
		arguments = append(arguments, "--link-dest", options.LinkDest)
	}

	if options.Compress {
		arguments = append(arguments, "--compress")
	}

	if options.CompressLevel > 0 {
		arguments = append(arguments, "--compress-level", strconv.Itoa(options.CompressLevel))
	}

	if len(options.SkipCompress) > 0 {
		arguments = append(arguments, "--skip-compress", strings.Join(options.SkipCompress, ","))
	}

	if options.CVSExclude {
		arguments = append(arguments, "--cvs-exclude")
	}

	if options.Stats {
		arguments = append(arguments, "--stats")
	}

	if options.HumanReadable {
		arguments = append(arguments, "--human-readable")
	}

	if options.Progress {
		arguments = append(arguments, "--progress")
	}

	if options.PasswordFile != "" {
		arguments = append(arguments, "--password-file", options.PasswordFile)
	}

	if options.BandwidthLimit > 0 {
		arguments = append(arguments, "--bwlimit", strconv.Itoa(options.BandwidthLimit))
	}

	if options.IPv4 {
		arguments = append(arguments, "--ipv4")
	}

	if options.IPv6 {
		arguments = append(arguments, "--ipv6")
	}

	if options.Info != "" {
		arguments = append(arguments, "--info", options.Info)
	}

	if options.OutFormat {
		arguments = append(arguments, "--out-format=\"%n\"")
	}

	if len(options.Include) > 0 {
		for _, pattern := range options.Include {
			arguments = append(arguments, "--include="+pattern)
		}
	}

	if len(options.Exclude) > 0 {
		for _, pattern := range options.Exclude {
			arguments = append(arguments, "--exclude="+pattern)
		}
	}

	if options.Filter != "" {
		arguments = append(arguments, "--filter="+options.Filter)
	}

	if options.Chown != "" {
		arguments = append(arguments, "--chown="+options.Chown)
	}

	return arguments
}

func createDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func isExist(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
