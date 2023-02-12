import numpy as np
import os

# Load the weights from the mounted volume
# Take the location of the data from the environment variable, have to fix this later
storage_location = os.environ['DATASET_LOCATION']
split_location = os.environ['SPLIT_LOCATION']

x_train = np.load(f"{storage_location}/x_train.npy")
y_train = np.load(f"{storage_location}/y_train.npy")

# Divide the traininf data into n disjoint sets
n = int(os.environ['DIVISIONS'])
x_train = np.array_split(x_train, n)
y_train = np.array_split(y_train, n)

# Save the data to the mounted volume
for i in range(n):
    np.save(f"{split_location}/x_train_{i}.npy", x_train[i])
    np.save(f"{storage_location}/y_train_{i}.npy", y_train[i])