output "kaptain_role_arn" {
  value = "${aws_iam_role.kaptain_role.arn}"
}

output "sailor_role_arn" {
  value = "${aws_iam_role.sailor_role.arn}"
}
