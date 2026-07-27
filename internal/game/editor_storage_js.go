//go:build js

package game

import "syscall/js"

const (
	editorPackKey       = "community_nongrams.editor.pack"
	communityLibraryKey = "community_nongrams.community.library"
	communityProfileKey = "community_nongrams.community.profile"
	communityBioKey     = "community_nongrams.community.bio"
	communitySocialKey  = "community_nongrams.community.social"
	communityPaletteKey = "community_nongrams.community.palette"
	communityColorKey   = "community_nongrams.community.favorite_color"
	communityNameKey    = "community_nongrams.community.name"
)

func legacyStorageKey(suffix string) string {
	return "pix" + "aross" + suffix
}

func loadStorageValue(key, legacySuffix string) string {
	storage := js.Global().Get("localStorage")
	if storage.IsUndefined() || storage.IsNull() {
		return ""
	}
	value := storage.Call("getItem", key)
	if !value.IsUndefined() && !value.IsNull() {
		return value.String()
	}
	if legacySuffix == "" {
		return ""
	}
	value = storage.Call("getItem", legacyStorageKey(legacySuffix))
	if value.IsUndefined() || value.IsNull() {
		return ""
	}
	text := value.String()
	storage.Call("setItem", key, text)
	return text
}

func saveStorageValue(key, value string) bool {
	storage := js.Global().Get("localStorage")
	if storage.IsUndefined() || storage.IsNull() {
		return false
	}
	storage.Call("setItem", key, value)
	return true
}

func invokeJS(name string, args ...any) (js.Value, bool) {
	fn := js.Global().Get(name)
	if fn.IsUndefined() || fn.IsNull() {
		return js.Undefined(), false
	}
	return fn.Invoke(args...), true
}

func callJS(name string, args ...any) bool {
	_, ok := invokeJS(name, args...)
	return ok
}

func callJSString(name, fallback string, args ...any) string {
	value, ok := invokeJS(name, args...)
	if !ok || value.IsUndefined() || value.IsNull() {
		return fallback
	}
	return value.String()
}

func callJSBool(name string, args ...any) bool {
	value, ok := invokeJS(name, args...)
	return ok && !value.IsUndefined() && !value.IsNull() && value.Bool()
}

func saveCommunityProfile(raw string) bool {
	return saveStorageValue(communityProfileKey, raw)
}

func saveCommunityBio(bio string) bool {
	return saveStorageValue(communityBioKey, bio)
}

func loadCommunityBio() string {
	return loadStorageValue(communityBioKey, ".community.bio")
}

func saveCommunitySocial(social string) bool {
	return saveStorageValue(communitySocialKey, social)
}

func loadCommunitySocial() string {
	return loadStorageValue(communitySocialKey, ".community.social")
}

func saveCommunityPalette(palette string) bool {
	return saveStorageValue(communityPaletteKey, palette)
}

func loadCommunityPalette() string {
	return loadStorageValue(communityPaletteKey, ".community.palette")
}

func saveCommunityFavoriteColor(color string) bool {
	return saveStorageValue(communityColorKey, color)
}

func loadCommunityFavoriteColor() string {
	return loadStorageValue(communityColorKey, ".community.favorite_color")
}

func saveCommunityName(name string) bool {
	return saveStorageValue(communityNameKey, name)
}

func loadCommunityName() string {
	if value := loadStorageValue(communityNameKey, ".community.name"); value != "" {
		return value
	}
	return callJSString("communityDisplayName", "")
}

func takeTextPaste() string {
	return callJSString("takeTextPaste", "")
}

func loadCommunityProfile() string {
	return loadStorageValue(communityProfileKey, ".community.profile")
}

func communityAccountLabel() string {
	return callJSString("communityAccountLabel", "Sign in")
}

func communitySignedIn() bool {
	return callJSBool("communitySignedIn")
}

func saveCommunityData(raw string) bool {
	return saveStorageValue(communityLibraryKey, raw)
}

func loadCommunityData() string {
	return loadStorageValue(communityLibraryKey, ".community.library")
}

func requestCommunityImport(size int, vertical bool) bool {
	return callJS("requestCommunityImport", size, vertical)
}

func takeCommunityImport() string {
	return callJSString("takeCommunityImport", "")
}

func takeCommunityImportError() string {
	return callJSString("takeCommunityImportError", "")
}

func requestCommunitySignIn(email string) bool {
	return callJS("requestCommunitySignIn", email)
}

func requestCommunitySignOut() bool {
	return callJS("requestCommunitySignOut")
}

func requestCommunityGoogleSignIn() bool {
	return callJS("requestCommunityGoogleSignIn")
}

func requestCommunityPublish(raw string, submitOfficial, rightsConfirmed bool, preview string) bool {
	return callJS("requestCommunityPublish", raw, submitOfficial, rightsConfirmed, preview)
}

func requestCommunityPackPublish(raw, preview string) bool {
	return callJS("requestCommunityPackPublish", raw, preview)
}

func takeCommunityResult() string {
	return callJSString("takeCommunityResult", "")
}

func takeCommunityPublishedID() string {
	return callJSString("takeCommunityPublishedID", "")
}

func takeCommunityPublishedPackID() string {
	return callJSString("takeCommunityPublishedPackID", "")
}

func requestCommunityCatalog(kind string) bool {
	return callJS("requestCommunityCatalog", kind)
}

func takeCommunityCatalog() string {
	return callJSString("takeCommunityCatalog", "")
}

func syncCommunityDraft(raw string) {
	callJS("syncCommunityDraft", raw)
}

func requestCommunityCloudDrafts() bool {
	return callJS("requestCommunityCloudDrafts")
}

func takeCommunityCloudDrafts() string {
	return callJSString("takeCommunityCloudDrafts", "")
}

func requestCommunityCreators() bool {
	return callJS("requestCommunityCreators")
}

func takeCommunityCreators() string {
	return callJSString("takeCommunityCreators", "")
}

func syncCommunityProfile(raw, bio, name, social, palette, favoriteColor string) {
	callJS("syncCommunityProfile", raw, bio, name, social, palette, favoriteColor)
}

func deleteCommunityCloudDraft(id string) {
	callJS("deleteCommunityCloudDraft", id)
}

func requestCommunityGallery(kind, sort string) bool {
	return callJS("requestCommunityGallery", kind, sort)
}

func takeCommunityGallery() string {
	return callJSString("takeCommunityGallery", "")
}

func requestCommunityChat(kind, id string) bool {
	return callJS("requestCommunityChat", kind, id)
}

func takeCommunityChat() string {
	return callJSString("takeCommunityChat", "")
}

func postCommunityChat(kind, id, body string) bool {
	return callJS("postCommunityChat", kind, id, body)
}

func recordCommunityPlay(levelID string, completed bool) {
	callJS("recordCommunityPlay", levelID, completed)
}

func requestCommunityCompleted() bool {
	return callJS("requestCommunityCompleted")
}

func takeCommunityCompleted() string {
	return callJSString("takeCommunityCompleted", "")
}

func requestCommunityPublished() bool {
	return callJS("requestCommunityPublished")
}

func takeCommunityPublished() string {
	return callJSString("takeCommunityPublished", "")
}

func unpublishCommunityItem(kind, id string) bool {
	return callJS("unpublishCommunityItem", kind, id)
}

func unpublishCommunityLocalArt(id string) bool {
	return callJS("unpublishCommunityLocalArt", id)
}

func updateCommunityPublishedItem(kind, id, title, description, levelsRaw string) bool {
	return callJS("updateCommunityPublishedItem", kind, id, title, description, levelsRaw)
}

func toggleCommunityLike(kind, id string) bool {
	return callJS("toggleCommunityLike", kind, id)
}

func promoteCommunityItem(kind, id string) bool {
	return callJS("promoteCommunityItem", kind, id)
}

func saveEditorPack(raw string) bool {
	return saveStorageValue(editorPackKey, raw)
}

func loadEditorPack() string {
	return loadStorageValue(editorPackKey, "")
}

func exportEditorImage(filename, raw string) bool {
	return callJS("downloadEditorImage", filename, raw)
}

func requestEditorImageImport(size int) bool {
	return callJS("requestEditorImageImport", size)
}

func takeEditorImageImport() string {
	return callJSString("takeEditorImageImport", "")
}

func requestEditorColorPicker(initial string) bool {
	return callJS("requestEditorColorPicker", initial)
}

func takeEditorColorPicker() string {
	return callJSString("takeEditorColorPicker", "")
}

func requestCommunityCoverImport(size int) bool {
	return callJS("requestCommunityCoverImport", size)
}

func takeCommunityCoverImport() string {
	return callJSString("takeCommunityCoverImport", "")
}

func requestEditorPackImport() bool {
	return callJS("requestEditorPackImport")
}

func takeEditorPackImport() string {
	return callJSString("takeEditorPackImport", "")
}

func communityFetchStatus() string {
	return callJSString("communityFetchStatus", "Supabase not configured")
}
