$BINARY = "./blockchain_go.exe"
						function startNode {
							 Write-Host "====>startNode"
						  & $BINARY startnode
						}
					Set-Item -Path "env:NODE_ID" -Value "192.168.254.129 3004"
					startNode