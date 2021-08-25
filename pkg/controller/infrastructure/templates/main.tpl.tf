terraform {
  required_providers {
    metal = {
      source  = "equinix/metal"
    }
  }
}

provider "metal" {
  auth_token = var.EQXM_API_KEY
}

resource "metal_project_ssh_key" "publickey" {
  name       = "{{ .clusterName }}-ssh-publickey"
  public_key = "{{ .sshPublicKey }}"
  project_id = var.EQXM_PROJECT_ID
}

output "{{ .outputKeys.sshKeyID }}" {
  value = metal_project_ssh_key.publickey.id
}
