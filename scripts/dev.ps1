param([ValidateSet("up","down","logs","check","migrate","test","reset")][string]$Action = "up")
$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)
function Invoke-Docker {
 param([string[]]$DockerArgs)
 & docker @DockerArgs
 if ($LASTEXITCODE -ne 0) { throw "Docker command failed: $DockerArgs" }
}
switch ($Action) {
 "up" { Invoke-Docker -DockerArgs @("compose","up","-d","--build") }
 "down" { Invoke-Docker -DockerArgs @("compose","down") }
 "logs" { Invoke-Docker -DockerArgs @("compose","logs","-f","--tail","100") }
 "check" { Invoke-Docker -DockerArgs @("compose","config","--quiet") }
 "migrate" {
  foreach ($service in @("user","commerce","wallet")) {
   Invoke-Docker -DockerArgs @("compose","run","--rm","$service-migrate","up")
  }
 }
 "test" {
  & go test ./...
  if ($LASTEXITCODE -ne 0) { throw "Go tests failed" }
  & npm --prefix web/portal run typecheck
  if ($LASTEXITCODE -ne 0) { throw "Typecheck failed" }
 }
 "reset" {
  $answer = Read-Host "Delete ALL local Supay database and NATS data? Type RESET"
  if ($answer -cne "RESET") { throw "Reset cancelled" }
  Invoke-Docker -DockerArgs @("compose","down","--volumes","--remove-orphans")
 }
}
