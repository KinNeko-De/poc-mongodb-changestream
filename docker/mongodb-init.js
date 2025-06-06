// This script runs during MongoDB container initialization
print('Starting MongoDB database initialization...');

// Create the application database
db = db.getSiblingDB('changestream_poc');

// Create the test_data collection if it doesn't exist
if (!db.getCollectionNames().includes('test_data')) {
  db.createCollection('test_data');
  print('Created collection: test_data');
} else {
  print('Collection test_data already exists');
}

// Create a unique index on the 'name' field
db.test_data.createIndex({ 'name': 1 }, { unique: true });
print('Created/Verified unique index on name field');

// Show the collection and index status
print('Collections:');
db.getCollectionNames().forEach(c => print(' - ' + c));
print('Indexes on test_data:');
db.test_data.getIndexes().forEach(idx => printjson(idx));

print('MongoDB database initialization completed successfully!');
