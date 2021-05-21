provider "packet" {
  auth_token = "${var.PACKET_API_KEY}"
}

// Deploy a new ssh key
resource "packet_project_ssh_key" "publickey" {
  name = "{{ .clusterName }}-ssh-publickey"
  public_key = "{{ .sshPublicKey }}"
  project_id = "{{ .packet.projectID }}"
}

//=====================================================================
//= Output variables
//=====================================================================

output "{{ .outputKeys.sshKeyID }}" {
  value = "${packet_project_ssh_key.publickey.id}"
}
