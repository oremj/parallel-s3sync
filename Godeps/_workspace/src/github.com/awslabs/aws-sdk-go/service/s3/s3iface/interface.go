package s3iface

import (
	"github.com/oremj/parallel-s3sync/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/service/s3"
)

type S3API interface {
	AbortMultipartUpload(*s3.AbortMultipartUploadInput) (*s3.AbortMultipartUploadOutput, error)

	CompleteMultipartUpload(*s3.CompleteMultipartUploadInput) (*s3.CompleteMultipartUploadOutput, error)

	CopyObject(*s3.CopyObjectInput) (*s3.CopyObjectOutput, error)

	CreateBucket(*s3.CreateBucketInput) (*s3.CreateBucketOutput, error)

	CreateMultipartUpload(*s3.CreateMultipartUploadInput) (*s3.CreateMultipartUploadOutput, error)

	DeleteBucket(*s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)

	DeleteBucketCORS(*s3.DeleteBucketCORSInput) (*s3.DeleteBucketCORSOutput, error)

	DeleteBucketLifecycle(*s3.DeleteBucketLifecycleInput) (*s3.DeleteBucketLifecycleOutput, error)

	DeleteBucketPolicy(*s3.DeleteBucketPolicyInput) (*s3.DeleteBucketPolicyOutput, error)

	DeleteBucketReplication(*s3.DeleteBucketReplicationInput) (*s3.DeleteBucketReplicationOutput, error)

	DeleteBucketTagging(*s3.DeleteBucketTaggingInput) (*s3.DeleteBucketTaggingOutput, error)

	DeleteBucketWebsite(*s3.DeleteBucketWebsiteInput) (*s3.DeleteBucketWebsiteOutput, error)

	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)

	DeleteObjects(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)

	GetBucketACL(*s3.GetBucketACLInput) (*s3.GetBucketACLOutput, error)

	GetBucketCORS(*s3.GetBucketCORSInput) (*s3.GetBucketCORSOutput, error)

	GetBucketLifecycle(*s3.GetBucketLifecycleInput) (*s3.GetBucketLifecycleOutput, error)

	GetBucketLocation(*s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error)

	GetBucketLogging(*s3.GetBucketLoggingInput) (*s3.GetBucketLoggingOutput, error)

	GetBucketNotification(*s3.GetBucketNotificationInput) (*s3.GetBucketNotificationOutput, error)

	GetBucketPolicy(*s3.GetBucketPolicyInput) (*s3.GetBucketPolicyOutput, error)

	GetBucketReplication(*s3.GetBucketReplicationInput) (*s3.GetBucketReplicationOutput, error)

	GetBucketRequestPayment(*s3.GetBucketRequestPaymentInput) (*s3.GetBucketRequestPaymentOutput, error)

	GetBucketTagging(*s3.GetBucketTaggingInput) (*s3.GetBucketTaggingOutput, error)

	GetBucketVersioning(*s3.GetBucketVersioningInput) (*s3.GetBucketVersioningOutput, error)

	GetBucketWebsite(*s3.GetBucketWebsiteInput) (*s3.GetBucketWebsiteOutput, error)

	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)

	GetObjectACL(*s3.GetObjectACLInput) (*s3.GetObjectACLOutput, error)

	GetObjectTorrent(*s3.GetObjectTorrentInput) (*s3.GetObjectTorrentOutput, error)

	HeadBucket(*s3.HeadBucketInput) (*s3.HeadBucketOutput, error)

	HeadObject(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error)

	ListBuckets(*s3.ListBucketsInput) (*s3.ListBucketsOutput, error)

	ListMultipartUploads(*s3.ListMultipartUploadsInput) (*s3.ListMultipartUploadsOutput, error)

	ListObjectVersions(*s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error)

	ListObjects(*s3.ListObjectsInput) (*s3.ListObjectsOutput, error)

	ListParts(*s3.ListPartsInput) (*s3.ListPartsOutput, error)

	PutBucketACL(*s3.PutBucketACLInput) (*s3.PutBucketACLOutput, error)

	PutBucketCORS(*s3.PutBucketCORSInput) (*s3.PutBucketCORSOutput, error)

	PutBucketLifecycle(*s3.PutBucketLifecycleInput) (*s3.PutBucketLifecycleOutput, error)

	PutBucketLogging(*s3.PutBucketLoggingInput) (*s3.PutBucketLoggingOutput, error)

	PutBucketNotification(*s3.PutBucketNotificationInput) (*s3.PutBucketNotificationOutput, error)

	PutBucketPolicy(*s3.PutBucketPolicyInput) (*s3.PutBucketPolicyOutput, error)

	PutBucketReplication(*s3.PutBucketReplicationInput) (*s3.PutBucketReplicationOutput, error)

	PutBucketRequestPayment(*s3.PutBucketRequestPaymentInput) (*s3.PutBucketRequestPaymentOutput, error)

	PutBucketTagging(*s3.PutBucketTaggingInput) (*s3.PutBucketTaggingOutput, error)

	PutBucketVersioning(*s3.PutBucketVersioningInput) (*s3.PutBucketVersioningOutput, error)

	PutBucketWebsite(*s3.PutBucketWebsiteInput) (*s3.PutBucketWebsiteOutput, error)

	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)

	PutObjectACL(*s3.PutObjectACLInput) (*s3.PutObjectACLOutput, error)

	RestoreObject(*s3.RestoreObjectInput) (*s3.RestoreObjectOutput, error)

	UploadPart(*s3.UploadPartInput) (*s3.UploadPartOutput, error)

	UploadPartCopy(*s3.UploadPartCopyInput) (*s3.UploadPartCopyOutput, error)
}