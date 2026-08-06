//go:build !js

package game

func saveCommunityProfile(string) bool { return false }

func loadCommunityProfile() string { return "" }

func saveCommunityBio(string) bool { return false }

func loadCommunityBio() string { return "" }

func saveCommunitySocial(string) bool { return false }

func loadCommunitySocial() string { return "" }

func saveCommunityPalette(string) bool { return false }

func loadCommunityPalette() string { return "" }

func saveCommunityFavoriteColor(string) bool { return false }

func loadCommunityFavoriteColor() string { return "" }

func saveCommunityName(string) bool { return false }

func loadCommunityName() string { return "" }

func takeTextPaste() string { return "" }

func communityAccountLabel() string { return "Sign in" }

func communitySignedIn() bool { return false }

func saveCommunityData(string) bool { return false }

func loadCommunityData() string { return "" }

func requestCommunityImport(int, bool, bool) bool { return false }

func takeCommunityImport() string { return "" }

func takeCommunityImportError() string { return "" }

func requestCommunitySignIn(string) bool { return false }

func requestCommunitySignOut() bool { return false }

func requestCommunityGoogleSignIn() bool { return false }

func requestCommunityPublish(string, bool, bool, string) bool { return false }

func requestCommunityPackPublish(string, string) bool { return false }

func takeCommunityResult() string { return "" }

func takeCommunityPublishedID() string { return "" }

func takeCommunityPublishedPackID() string { return "" }

func requestCommunityCatalog(string) bool { return false }

func takeCommunityCatalog() string { return "" }

func syncCommunityDraft(string) {}

func deleteCommunityCloudDraft(string) {}

func requestCommunityCloudDrafts() bool { return false }

func takeCommunityCloudDrafts() string { return "" }

func requestCommunityCreators() bool { return false }

func takeCommunityCreators() string { return "" }

func syncCommunityProfile(string, string, string, string, string, string) {}

func syncCommunityProfileDetails(string, string, string, string, string) {}

func requestCommunityGallery(string, string) bool { return false }

func takeCommunityGallery() string { return "" }

func requestCommunityChat(string, string) bool { return false }

func takeCommunityChat() string { return "" }

func postCommunityChat(string, string, string) bool { return false }

func recordCommunityPlay(string, bool) {}

func requestCommunityCompleted() bool { return false }

func takeCommunityCompleted() string { return "" }

func requestCommunityPublished() bool { return false }

func takeCommunityPublished() string { return "" }

func unpublishCommunityItem(string, string) bool { return false }

func unpublishCommunityLocalArt(string) bool { return false }

func updateCommunityPublishedItem(string, string, string, string, string) bool { return false }

func toggleCommunityLike(string, string) bool { return false }

func promoteCommunityItem(string, string) bool { return false }

func saveEditorPack(string) bool {
	return false
}

func loadEditorPack() string {
	return ""
}

func exportEditorImage(string, string) bool {
	return false
}

func requestEditorImageImport(int) bool {
	return false
}

func takeEditorImageImport() string {
	return ""
}

func requestEditorColorPicker(string) bool {
	return false
}

func takeEditorColorPicker() string {
	return ""
}

func requestCommunityCoverImport(int) bool { return false }

func takeCommunityCoverImport() string { return "" }

func requestEditorPackImport() bool {
	return false
}

func takeEditorPackImport() string {
	return ""
}

func communityFetchStatus() string {
	return "Community is available in the web build"
}
