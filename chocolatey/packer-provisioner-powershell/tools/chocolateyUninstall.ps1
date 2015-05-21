$unzipLocation = "$env:SystemDrive\HashiCorp\packer"
$pluginExe = $unzipLocation + "\packer-provisioner-powershell.exe"

if (Test-Path $pluginExe) {
  Remove-Item $pluginExe
}