$unzipLocation = "$env:SystemDrive\HashiCorp\packer"
$pluginExe = $unzipLocation + "\packer-builder-hyperv-iso.exe"

if (Test-Path $pluginExe) {
  Remove-Item $pluginExe
}