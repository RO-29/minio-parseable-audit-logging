const Minio = require('minio');
const { faker } = require('@faker-js/faker');
const fs = require('fs');
const path = require('path');

// MinIO client configuration
const minioClient = new Minio.Client({
    endPoint: 'localhost',
    port: 9000,
    useSSL: false,
    accessKey: 'minioadmin',
    secretKey: 'minioadmin'
});

const bucketName = 'js-test-bucket';

// Ensure local directories exist
const uploadDir = './uploads';
const downloadDir = './downloads';

if (!fs.existsSync(uploadDir)) {
    fs.mkdirSync(uploadDir, { recursive: true });
}

if (!fs.existsSync(downloadDir)) {
    fs.mkdirSync(downloadDir, { recursive: true });
}

async function createBucketIfNotExists() {
    try {
        console.log(`🔍 Checking if bucket '${bucketName}' exists...`);
        const exists = await minioClient.bucketExists(bucketName);
        if (!exists) {
            console.log(`📦 Creating bucket '${bucketName}'...`);
            await minioClient.makeBucket(bucketName, 'us-east-1');
            console.log(`✅ Bucket '${bucketName}' created successfully`);
        } else {
            console.log(`✅ Bucket '${bucketName}' already exists`);
        }
        
        // Set bucket policy for better logging
        const policy = {
            Version: '2012-10-17',
            Statement: [{
                Effect: 'Allow',
                Principal: { AWS: ['*'] },
                Action: ['s3:GetObject'],
                Resource: [`arn:aws:s3:::${bucketName}/*`]
            }]
        };
        
        try {
            await minioClient.setBucketPolicy(bucketName, JSON.stringify(policy));
            console.log(`🔒 Set bucket policy for ${bucketName}`);
        } catch (policyError) {
            console.log(`⚠️  Could not set bucket policy: ${policyError.message}`);
        }
        
    } catch (error) {
        console.error('❌ Error with bucket operations:', error);
        throw error;
    }
}

function generateRandomFile() {
    const fileName = `${faker.system.fileName()}.txt`;
    const fileSize = faker.number.int({ min: 1000, max: 50000 }); // 1KB to 50KB
    const content = faker.lorem.paragraphs(Math.ceil(fileSize / 100));
    
    const filePath = path.join(uploadDir, fileName);
    fs.writeFileSync(filePath, content);
    
    return { fileName, filePath, size: content.length };
}

async function uploadFile(fileName, filePath) {
    try {
        const stats = fs.statSync(filePath);
        console.log(`📤 Uploading ${fileName} (${stats.size} bytes)...`);
        
        // Add metadata to trigger more audit log details
        const metaData = {
            'Content-Type': 'text/plain',
            'X-Demo-App': 'parseable-minio-demo',
            'X-File-Size': stats.size.toString(),
            'X-Upload-Time': new Date().toISOString()
        };
        
        await minioClient.fPutObject(bucketName, fileName, filePath, metaData);
        console.log(`✅ Successfully uploaded ${fileName}`);
        
        // Delete local file after upload
        fs.unlinkSync(filePath);
        console.log(`🗑️  Deleted local file ${fileName}`);
        
        return stats.size;
    } catch (error) {
        console.error(`❌ Error uploading ${fileName}:`, error);
        throw error;
    }
}

async function downloadFile(fileName) {
    try {
        const downloadPath = path.join(downloadDir, `downloaded_${fileName}`);
        console.log(`📥 Downloading ${fileName}...`);
        
        await minioClient.fGetObject(bucketName, fileName, downloadPath);
        console.log(`✅ Successfully downloaded ${fileName} to ${downloadPath}`);
        
        const stats = fs.statSync(downloadPath);
        return stats.size;
    } catch (error) {
        console.error(`❌ Error downloading ${fileName}:`, error);
        throw error;
    }
}

async function listObjects() {
    try {
        console.log(`📋 Listing objects in bucket ${bucketName}:`);
        const objectsList = [];
        
        const stream = minioClient.listObjects(bucketName, '', true);
        
        return new Promise((resolve, reject) => {
            stream.on('data', (obj) => {
                objectsList.push(obj);
                console.log(`   - ${obj.name} (${obj.size} bytes, modified: ${obj.lastModified})`);
            });
            
            stream.on('error', (err) => {
                reject(err);
            });
            
            stream.on('end', () => {
                resolve(objectsList);
            });
        });
    } catch (error) {
        console.error('❌ Error listing objects:', error);
        throw error;
    }
}

async function performAdditionalOperations() {
    try {
        console.log('🔧 Performing additional MinIO operations for audit logs...');
        
        // Get bucket info
        console.log('📊 Getting bucket information...');
        try {
            await minioClient.getBucketPolicy(bucketName);
            console.log('✅ Retrieved bucket policy');
        } catch (e) {
            console.log('⚠️  No bucket policy found (this is normal)');
        }
        
        // List objects with prefix
        console.log('🔍 Listing objects with different parameters...');
        const prefixStream = minioClient.listObjects(bucketName, 'sample', false);
        let prefixCount = 0;
        
        await new Promise((resolve, reject) => {
            prefixStream.on('data', (obj) => {
                prefixCount++;
            });
            prefixStream.on('error', reject);
            prefixStream.on('end', resolve);
        });
        
        console.log(`📋 Found ${prefixCount} objects with 'sample' prefix`);
        
        // Try to access a non-existent object (will generate 404 audit log)
        try {
            await minioClient.getObject(bucketName, 'non-existent-file.txt');
        } catch (error) {
            console.log('🔍 Attempted to access non-existent file (generates 404 audit log)');
        }
        
        // Get bucket location
        try {
            const location = await minioClient.getBucketLocation(bucketName);
            console.log(`🌍 Bucket location: ${location}`);
        } catch (e) {
            console.log('⚠️  Could not get bucket location');
        }
        
    } catch (error) {
        console.error('❌ Error in additional operations:', error);
    }
}

async function runDemo() {
    try {
        console.log('🚀 Starting Parseable + MinIO Demo Application');
        console.log('=' .repeat(50));
        
        // Create bucket if it doesn't exist
        await createBucketIfNotExists();
        
        // Generate random number of files (1-5)
        const numFiles = faker.number.int({ min: 1, max: 5 });
        console.log(`📁 Generating ${numFiles} random files...`);
        
        const files = [];
        for (let i = 0; i < numFiles; i++) {
            const file = generateRandomFile();
            files.push(file);
            console.log(`   ${i + 1}. ${file.fileName} (${file.size} bytes)`);
        }
        
        console.log('');
        
        // Upload all files
        for (const file of files) {
            await uploadFile(file.fileName, file.filePath);
            // Add small delay to see audit logs separately
            await new Promise(resolve => setTimeout(resolve, 500));
        }
        
        console.log('');
        
        // List all objects in bucket
        await listObjects();
        
        console.log('');
        
        // Download some files back
        const filesToDownload = files.slice(0, Math.min(2, files.length));
        for (const file of filesToDownload) {
            await downloadFile(file.fileName);
            await new Promise(resolve => setTimeout(resolve, 500));
        }
        
        console.log('');
        
        // Perform additional operations to generate more audit logs
        await performAdditionalOperations();
        
        console.log('');
        console.log('✅ Demo completed successfully!');
        console.log('🔍 Check your Parseable dashboard at http://localhost:8000');
        console.log('   - Stream: minio_audit (for MinIO audit logs)');
        console.log('   - Stream: minio_log (for MinIO server logs)');
        console.log('🗂️  Check your MinIO console at http://localhost:9001');
        
    } catch (error) {
        console.error('❌ Demo failed:', error);
        process.exit(1);
    }
}

// Run demo if this file is executed directly
if (require.main === module) {
    runDemo();
}

module.exports = { runDemo };
