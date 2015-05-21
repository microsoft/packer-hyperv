$url = 'https://github.com/MSOpenTech/packer-hyperv/raw/master/bin/packer-provisioner-powershell.exe'
$unzipLocation = "$env:SystemDrive\HashiCorp\packer\packer-provisioner-powershell.exe"

Get-ChocolateyWebFile "packer-provisioner-powershell" $unzipLocation $url 