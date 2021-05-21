provider "packet" {
  auth_token = "${var.PACKET_API_KEY}"
}

// Deploy a new ssh key
resource "packet_project_ssh_key" "publickey" {
  name = "{{ .Values.clusterName }}-ssh-publickey"
  public_key = "{{ .Values.sshPublicKey }}"
  project_id = "{{ .Values.packet.projectID }}"
}

//=====================================================================
//= Output variables
//=====================================================================

output "{{ .Values.outputKeys.sshKeyID }}" {
  value = "${packet_project_ssh_key.publickey.id}"
}
