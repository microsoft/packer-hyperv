Write-Host "Starting file script..." 

Write-Host 'Executing [DateTime]::Now...' 
[DateTime]::Now

Write-Host 'Executing Install-WindowsFeature -Name "XPS-Viewer" -IncludeAllSubFeature' 
Install-WindowsFeature -Name "XPS-Viewer" -IncludeAllSubFeature

Write-Host "File script finished!" 
