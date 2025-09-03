<template>
  <div class="storage-config">
    <h3>Storage Configuration</h3>

    <!-- Storage Type Selector -->
    <div class="form-group">
      <label for="storageType">Storage Type</label>
      <select id="storageType" v-model="config.storageType" class="form-input">
        <option value="local">Local Storage</option>
        <option value="s3">Amazon S3 / S3-Compatible</option>
        <option value="gcs">Google Cloud Storage</option>
        <option value="azure">Azure Blob Storage</option>
      </select>
      <p class="form-help">Choose where Homebox will store your data</p>
    </div>

    <!-- Local Storage Configuration -->
    <div v-if="config.storageType === 'local'" class="storage-section">
      <h4>Local Storage Settings</h4>

      <div class="form-group">
        <label for="localType">Storage Type</label>
        <select id="localType" v-model="config.storageConfig.local.type" class="form-input">
          <option value="volume">Docker Volume</option>
          <option value="directory">Host Directory</option>
        </select>
      </div>

      <div v-if="config.storageConfig.local.type === 'directory'" class="form-group">
        <label for="localDirectory">Host Directory Path</label>
        <input
          id="localDirectory"
          v-model="config.storageConfig.local.directory"
          type="text"
          class="form-input"
          placeholder="./homebox-data"
        />
        <p class="form-help">Path on the host system where data will be stored</p>
      </div>

      <div v-if="config.storageConfig.local.type === 'volume'" class="form-group">
        <label for="localVolume">Volume Name</label>
        <input
          id="localVolume"
          v-model="config.storageConfig.local.volumeName"
          type="text"
          class="form-input"
          placeholder="homebox-data"
        />
      </div>

      <div class="form-group">
        <label for="localPath">Custom Storage Path (Optional)</label>
        <input
          id="localPath"
          v-model="config.storageConfig.local.path"
          type="text"
          class="form-input"
          placeholder="/data"
        />
        <p class="form-help">Custom path inside the container. Leave as /data for default.</p>
      </div>
    </div>

    <!-- S3 Storage Configuration -->
    <div v-if="config.storageType === 's3'" class="storage-section">
      <h4>S3 Storage Settings</h4>

      <div class="form-group">
        <label>
          <input
            type="checkbox"
            v-model="config.storageConfig.s3.isCompatible"
            class="form-checkbox"
          />
          Use S3-Compatible Storage (MinIO, Cloudflare R2, Backblaze B2, etc.)
        </label>
      </div>

      <div v-if="config.storageConfig.s3.isCompatible" class="form-group">
        <label for="s3Service">S3-Compatible Service</label>
        <select id="s3Service" v-model="config.storageConfig.s3.compatibleService" class="form-input">
          <option value="">Custom/Other</option>
          <option value="minio">MinIO</option>
          <option value="cloudflare-r2">Cloudflare R2</option>
          <option value="backblaze-b2">Backblaze B2</option>
        </select>
      </div>

      <div class="form-group">
        <label for="s3Bucket">Bucket Name</label>
        <input
          id="s3Bucket"
          v-model="config.storageConfig.s3.bucket"
          type="text"
          class="form-input"
          placeholder="my-homebox-bucket"
          required
        />
      </div>

      <div v-if="!config.storageConfig.s3.isCompatible" class="form-group">
        <label for="s3Region">AWS Region</label>
        <input
          id="s3Region"
          v-model="config.storageConfig.s3.region"
          type="text"
          class="form-input"
          placeholder="us-east-1"
          required
        />
      </div>

      <div v-if="config.storageConfig.s3.isCompatible" class="form-group">
        <label for="s3Endpoint">Endpoint URL</label>
        <input
          id="s3Endpoint"
          v-model="config.storageConfig.s3.endpoint"
          type="text"
          class="form-input"
          :placeholder="getS3EndpointPlaceholder()"
        />
        <p class="form-help">The endpoint URL for your S3-compatible service</p>
      </div>

      <div class="form-group">
        <label for="s3AccessKey">AWS Access Key ID</label>
        <input
          id="s3AccessKey"
          v-model="config.storageConfig.s3.awsAccessKeyId"
          type="text"
          class="form-input"
          placeholder="AKIAIOSFODNN7EXAMPLE"
          required
        />
      </div>

      <div class="form-group">
        <label for="s3SecretKey">AWS Secret Access Key</label>
        <input
          id="s3SecretKey"
          v-model="config.storageConfig.s3.awsSecretAccessKey"
          type="password"
          class="form-input"
          placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
          required
        />
      </div>

      <div class="form-group">
        <label for="s3SessionToken">AWS Session Token (Optional)</label>
        <input
          id="s3SessionToken"
          v-model="config.storageConfig.s3.awsSessionToken"
          type="password"
          class="form-input"
          placeholder="For temporary credentials"
        />
        <p class="form-help">Only needed for temporary AWS credentials</p>
      </div>

      <div class="form-group">
        <label for="s3PrefixPath">Storage Prefix Path (Optional)</label>
        <input
          id="s3PrefixPath"
          v-model="config.storageConfig.s3.prefixPath"
          type="text"
          class="form-input"
          placeholder="homebox/"
        />
        <p class="form-help">Prefix for all stored objects in the bucket</p>
      </div>

      <!-- Advanced S3 Settings -->
      <details class="advanced-settings">
        <summary>Advanced S3 Settings</summary>

        <div class="form-group">
          <label for="s3AwsSdk">AWS SDK Version</label>
          <select id="s3AwsSdk" v-model="config.storageConfig.s3.awsSdk" class="form-input">
            <option value="v2">v2 (Recommended)</option>
            <option value="v1">v1</option>
          </select>
        </div>

        <div class="form-group">
          <label>
            <input
              type="checkbox"
              v-model="config.storageConfig.s3.disableSSL"
              class="form-checkbox"
            />
            Disable SSL
          </label>
        </div>

        <div class="form-group">
          <label>
            <input
              type="checkbox"
              v-model="config.storageConfig.s3.s3ForcePathStyle"
              class="form-checkbox"
            />
            Force Path Style Access
          </label>
        </div>

        <div class="form-group">
          <label for="s3SseType">Server-Side Encryption</label>
          <select id="s3SseType" v-model="config.storageConfig.s3.sseType" class="form-input">
            <option value="">None</option>
            <option value="AES256">AES256</option>
            <option value="aws:kms">AWS KMS</option>
            <option value="aws:kms:dsse">AWS KMS DSSE</option>
          </select>
        </div>

        <div v-if="config.storageConfig.s3.sseType.includes('kms')" class="form-group">
          <label for="s3KmsKey">KMS Key ID</label>
          <input
            id="s3KmsKey"
            v-model="config.storageConfig.s3.kmsKeyId"
            type="text"
            class="form-input"
            placeholder="arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
          />
        </div>

        <div class="form-group">
          <label>
            <input
              type="checkbox"
              v-model="config.storageConfig.s3.fips"
              class="form-checkbox"
            />
            Use FIPS Endpoints
          </label>
        </div>

        <div class="form-group">
          <label>
            <input
              type="checkbox"
              v-model="config.storageConfig.s3.dualstack"
              class="form-checkbox"
            />
            Use Dual-Stack Endpoints
          </label>
        </div>

        <div class="form-group">
          <label>
            <input
              type="checkbox"
              v-model="config.storageConfig.s3.accelerate"
              class="form-checkbox"
            />
            Use S3 Transfer Acceleration
          </label>
        </div>
      </details>
    </div>

    <!-- Google Cloud Storage Configuration -->
    <div v-if="config.storageType === 'gcs'" class="storage-section">
      <h4>Google Cloud Storage Settings</h4>

      <div class="form-group">
        <label for="gcsBucket">Bucket Name</label>
        <input
          id="gcsBucket"
          v-model="config.storageConfig.gcs.bucket"
          type="text"
          class="form-input"
          placeholder="my-homebox-bucket"
          required
        />
      </div>

      <div class="form-group">
        <label for="gcsProject">Project ID</label>
        <input
          id="gcsProject"
          v-model="config.storageConfig.gcs.projectId"
          type="text"
          class="form-input"
          placeholder="my-gcp-project"
        />
      </div>

      <div class="form-group">
        <label for="gcsCredentialsPath">Service Account Key Path</label>
        <input
          id="gcsCredentialsPath"
          v-model="config.storageConfig.gcs.credentialsPath"
          type="text"
          class="form-input"
          placeholder="/app/gcs-credentials.json"
        />
        <p class="form-help">Path to the service account JSON key file inside the container</p>
      </div>

      <div class="form-group">
        <label for="gcsPrefixPath">Storage Prefix Path (Optional)</label>
        <input
          id="gcsPrefixPath"
          v-model="config.storageConfig.gcs.prefixPath"
          type="text"
          class="form-input"
          placeholder="homebox/"
        />
        <p class="form-help">Prefix for all stored objects in the bucket</p>
      </div>

      <div class="info-box">
        <h5>📋 Setup Instructions:</h5>
        <ol>
          <li>Create a service account in your GCP project</li>
          <li>Grant Storage Admin permissions to the service account</li>
          <li>Download the JSON key file</li>
          <li>Mount the key file as a read-only volume in your container</li>
          <li>Set GOOGLE_APPLICATION_CREDENTIALS environment variable</li>
        </ol>
      </div>
    </div>

    <!-- Azure Blob Storage Configuration -->
    <div v-if="config.storageType === 'azure'" class="storage-section">
      <h4>Azure Blob Storage Settings</h4>

      <div class="form-group">
        <label>
          <input
            type="checkbox"
            v-model="config.storageConfig.azure.useEmulator"
            class="form-checkbox"
          />
          Use Azure Storage Emulator (for development)
        </label>
      </div>

      <div class="form-group">
        <label for="azureContainer">Container Name</label>
        <input
          id="azureContainer"
          v-model="config.storageConfig.azure.container"
          type="text"
          class="form-input"
          placeholder="homebox-container"
          required
        />
      </div>

      <div v-if="!config.storageConfig.azure.useEmulator" class="form-group">
        <label for="azureAccount">Storage Account Name</label>
        <input
          id="azureAccount"
          v-model="config.storageConfig.azure.storageAccount"
          type="text"
          class="form-input"
          placeholder="mystorageaccount"
          required
        />
      </div>

      <div v-if="!config.storageConfig.azure.useEmulator" class="form-group">
        <label for="azureKey">Storage Account Key</label>
        <input
          id="azureKey"
          v-model="config.storageConfig.azure.storageKey"
          type="password"
          class="form-input"
          placeholder="Your Azure storage account key"
          required
        />
      </div>

      <div v-if="!config.storageConfig.azure.useEmulator" class="form-group">
        <label for="azureSas">SAS Token (Optional)</label>
        <input
          id="azureSas"
          v-model="config.storageConfig.azure.sasToken"
          type="password"
          class="form-input"
          placeholder="?sv=2021-06-08&ss=b&srt=sco&sp=rwdlacupx&se=..."
        />
        <p class="form-help">Use SAS token instead of storage account key</p>
      </div>

      <div v-if="config.storageConfig.azure.useEmulator" class="form-group">
        <label for="azureEmulatorEndpoint">Emulator Endpoint</label>
        <input
          id="azureEmulatorEndpoint"
          v-model="config.storageConfig.azure.emulatorEndpoint"
          type="text"
          class="form-input"
          placeholder="localhost:10001"
        />
      </div>

      <div class="form-group">
        <label for="azurePrefixPath">Storage Prefix Path (Optional)</label>
        <input
          id="azurePrefixPath"
          v-model="config.storageConfig.azure.prefixPath"
          type="text"
          class="form-input"
          placeholder="homebox/"
        />
        <p class="form-help">Prefix for all stored objects in the container</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { defineProps } from 'vue'

const props = defineProps({
  config: {
    type: Object,
    required: true
  }
})

function getS3EndpointPlaceholder() {
  const service = props.config.storageConfig.s3.compatibleService
  switch (service) {
    case 'minio':
      return 'http://minio:9000'
    case 'cloudflare-r2':
      return 'https://<account-id>.r2.cloudflarestorage.com'
    case 'backblaze-b2':
      return 'https://s3.us-west-004.backblazeb2.com'
    default:
      return 'https://your-s3-compatible-endpoint.com'
  }
}
</script>

<style scoped>
.storage-config {
  padding: 1.5rem;
  background-color: var(--vp-c-bg-soft);
  border-radius: 8px;
}

.storage-section {
  margin-top: 1.5rem;
  padding: 1rem;
  background-color: var(--vp-c-bg);
  border-radius: 6px;
  border: 1px solid var(--vp-c-divider);
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.25rem;
  font-weight: 500;
  color: var(--vp-c-text-1);
}

.form-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-size: 0.875rem;
}

.form-input:focus {
  outline: none;
  border-color: var(--vp-c-brand);
  box-shadow: 0 0 0 2px var(--vp-c-brand-light);
}

.form-checkbox {
  width: auto;
  margin-right: 0.5rem;
}

.form-help {
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--vp-c-text-2);
}

.advanced-settings {
  margin-top: 1rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
}

.advanced-settings summary {
  padding: 0.75rem;
  background-color: var(--vp-c-bg-mute);
  cursor: pointer;
  font-weight: 500;
}

.advanced-settings[open] summary {
  border-bottom: 1px solid var(--vp-c-divider);
}

.advanced-settings .form-group {
  margin: 1rem;
}

.info-box {
  margin-top: 1rem;
  padding: 1rem;
  background-color: var(--vp-c-bg-alt);
  border-left: 4px solid var(--vp-c-brand);
  border-radius: 4px;
}

.info-box h5 {
  margin: 0 0 0.5rem 0;
  color: var(--vp-c-text-1);
}

.info-box ol {
  margin: 0;
  padding-left: 1.25rem;
}

.info-box li {
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
  color: var(--vp-c-text-2);
}

h3 {
  margin: 0 0 1.5rem 0;
  color: var(--vp-c-text-1);
  font-size: 1.25rem;
  font-weight: 600;
}

h4 {
  margin: 0 0 1rem 0;
  color: var(--vp-c-text-1);
  font-size: 1.1rem;
  font-weight: 600;
}
</style>