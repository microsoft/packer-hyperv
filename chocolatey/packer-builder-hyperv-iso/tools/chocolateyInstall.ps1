$url = 'https://github.com/MSOpenTech/packer-hyperv/raw/master/bin/packer-builder-hyperv-iso.exe'
$unzipLocation = "$env:SystemDrive\HashiCorp\packer\packer-builder-hyperv-iso.exe"

Get-ChocolateyWebFile "packer-builder-hyperv-iso" $unzipLocation $url 