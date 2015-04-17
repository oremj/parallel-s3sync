package s3sync

import "mime"

func init() {
	mime.AddExtensionType(".mar", "application/octet-stream")
	mime.AddExtensionType(".msi", "application/octet-stream")
	mime.AddExtensionType(".apk", "application/vnd.android.package-archive")
	mime.AddExtensionType(".cab", "application/vnd.ms-cab-compressed")
	mime.AddExtensionType(".dmg", "application/x-apple-diskimage")
	mime.AddExtensionType(".deb", "application/x-debian-package")
	mime.AddExtensionType(".install", "application/x-install-instructions")
	mime.AddExtensionType(".jar", "application/x-java-archive")
	mime.AddExtensionType(".xpi", "application/x-xpinstall")
}
