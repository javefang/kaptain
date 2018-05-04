resource "aws_iam_role" "kaptain_role" {
  name               = "${var.prefix}_${var.cluster_name}_kaptain_role"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role_policy.json}"
}

resource "aws_iam_role" "sailor_role" {
  name               = "${var.prefix}_${var.cluster_name}_sailor_role"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role_policy.json}"
}

resource "aws_iam_role_policy" "kaptain_policy" {
  name   = "KaptainPolicyReadWrite"
  role   = "${aws_iam_role.kaptain_role.id}"
  policy = "${data.aws_iam_policy_document.kaptain_policy.json}"
}

resource "aws_iam_role_policy" "sailor_policy" {
  name   = "KaptainPolicyReadOnly"
  role   = "${aws_iam_role.sailor_role.id}"
  policy = "${data.aws_iam_policy_document.sailor_policy.json}"
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${var.cluster_aws_account}:root"]
    }
  }
}

data "aws_iam_policy_document" "kaptain_policy" {
  statement {
    sid = "KaptainClusterReadWrite"

    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
    ]

    resources = [
      "arn:aws:s3:::${var.state_bucket}/${var.cluster_name}",
      "arn:aws:s3:::${var.state_bucket}/${var.cluster_name}/*",
    ]
  }

  statement {
    sid = "KaptainList"

    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
    ]

    resources = [
      "arn:aws:s3:::${var.state_bucket}",
    ]
  }
}

data "aws_iam_policy_document" "sailor_policy" {
  statement {
    sid = "KaptainClusterReadOnly"

    actions = [
      "s3:GetObject",
    ]

    resources = [
      "arn:aws:s3:::${var.state_bucket}/${var.cluster_name}",
      "arn:aws:s3:::${var.state_bucket}/${var.cluster_name}/*",
    ]
  }
}
