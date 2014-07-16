
Download the archive to your devbox and extract it into your packer folder.

The files builder-hyperv-iso.exe and provisioner-powershell.exe in the archive are Packer plugins. 

How to install plugins you can find here: http://www.packer.io/docs/extend/plugins.html

File _install.cmd can install the plugins for you.

The file packer-post-processor-vagrant.exe is an extention of the original file with the same name to create a vagrant box from a Hyper-V atrifact. 
