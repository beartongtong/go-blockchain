$sourceFilePath = "E:\blockBackup\79\blockchain_192.168.254.129 3000.db"
$destinationFolder = "E:\blockBackup\79"
$numberOfCopies = 9

for ($i = 1; $i -le $numberOfCopies; $i++) {
    $destinationPath = "E:\blockBackup\79\blockchain_192.168.254.129 300$i.db"

    Write-Host "Copying from: $sourceFilePath"
    Write-Host "Copying to:   $destinationPath"
    
    Copy-Item -Path "$sourceFilePath" -Destination "$destinationPath" -Force
}

Write-Host "end"
