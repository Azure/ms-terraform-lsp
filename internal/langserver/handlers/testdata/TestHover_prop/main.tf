resource "msgraph_resource" "group" {
  # url = "groups"
  url = "groups"
  body = {
    acceptedSenders = [
      {

      }
    ]
  }
}