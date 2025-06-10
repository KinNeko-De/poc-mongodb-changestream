print('Starting MongoDB replica set initialization...');

// Initialize replica set with localhost instead of mongodb
rs.initiate({
  _id: 'rs0',
  members: [
    { _id: 0, host: 'localhost:27017' }
  ]
});

// Wait for replica set to initialize
let attempts = 30;
while (attempts > 0) {
  try {
    const status = rs.status();
    if (status.ok && status.members && status.members[0].state === 1) {
      print('Replica set initialized successfully!');
      break;
    }
  } catch (e) {
    print('Error checking replica set status: ' + e);
  }
  print('Waiting for replica set to initialize... (' + attempts + ' attempts left)');
  sleep(1000);
  attempts--;
}

if (attempts === 0) {
  print('Failed to initialize replica set after multiple attempts');
  quit(1);
}

// Show replica set status
print('Final replica set status:');
printjson(rs.status());
