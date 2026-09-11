Write-Host "==> Compilando a imagem do Ansible Control Node..." -ForegroundColor Cyan
docker build -t korp-ansible-runner ./ansible

if ($LASTEXITCODE -ne 0) {
    Write-Error "Falha ao construir a imagem do Ansible Runner."
    exit 1
}

Write-Host "==> Executando a Playbook de Deploy via Ansible Containerizado..." -ForegroundColor Cyan
docker run --rm -it `
  -v //var/run/docker.sock:/var/run/docker.sock `
  -v "${PWD}:/workspace" `
  korp-ansible-runner playbooks/deploy.yml