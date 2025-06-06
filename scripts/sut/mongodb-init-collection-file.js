// This script runs during MongoDB container initialization
print('Starting MongoDB database initialization...');

// Create the application database
db = db.getSiblingDB('store_file');

// Create the file collection if it doesn't exist
if (!db.getCollectionNames().includes('file')) {
  db.createCollection('file');
  print('Created collection: file');
} else {
  print('Collection file already exists');
}

// Show the collection and index status
print('Collections:');
db.getCollectionNames().forEach(c => print(' - ' + c));
print('Indexes on file:');
db.file.getIndexes().forEach(idx => printjson(idx));

print('MongoDB database initialization completed successfully!');
