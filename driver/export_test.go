package driver

var (
	ExportGitSetUpAuth             = (*GitDriver).setUpAuth
	ExportGitIsPrivateKeyEncypted  = (*GitDriver).isPrivateKeyEncrypted
	ExportGitSetUpUsernamePassword = (*GitDriver).setUpUsernamePassword
	ExportGitSetUserInfo           = (*GitDriver).setUserInfo
	ExportGitSetupRepo             = (*GitDriver).setUpRepo
	ExportGitReadVersion           = (*GitDriver).readVersion
	ExportGitWriteVersion          = (*GitDriver).writeVersion

	ExportGitRepoDir        = gitRepoDir
	ExportGitPrivateKeyPATH = privateKeyPATH
	ExportGitNetRcPATH      = netRcPATH
)

func SetGitRepoDir(path string) (resetFunc func()) {
	var tmp string
	tmp, gitRepoDir = gitRepoDir, path
	return func() {
		gitRepoDir = tmp
	}
}
