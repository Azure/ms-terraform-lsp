resource "msgraph_resource" "group" {
  url = "groups"
  body = {
    acceptedSenders = [
      {
        deletedDateTime = ""
      }
    ]
  }
}