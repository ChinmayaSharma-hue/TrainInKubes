import tensorflow as tf
import numpy as np
import os
import pickle

# Loading the model from a persistent volume
# Take the location of the model from the environment variable, have to fix this later
# Hint - Use ConfigMaps

model_location = os.environ['MODEL_LOCATION']
gradient_location = os.environ['GRADIENT_LOCATION']
features_location = os.environ['FEATURES_LOCATION']
labels_location = os.environ['LABELS_LOCATION']
starting_index = int(os.environ['STARTING_INDEX'])
ending_index = int(os.environ['ENDING_INDEX'])
job_index = int(os.environ['JOB_INDEX'])

model = tf.keras.models.load_model(model_location)

# Loading the training data from a persistent volume
x_train = np.load(features_location)
y_train = np.load(labels_location)

# Get the training data for the current job
if ending_index > len(x_train):
    ending_index = len(x_train)

x_batch = x_train[starting_index:ending_index]
y_batch = y_train[starting_index:ending_index]

# Convert x_batch and y_batch to int_64t
x_batch = tf.convert_to_tensor(x_train, dtype=tf.float32)
y_batch = tf.convert_to_tensor(y_train, dtype=tf.int64)

# Print the shape of the training data
print(x_batch.shape)

## Define the Loss Function
loss_fn = tf.keras.losses.CategoricalCrossentropy()

with tf.GradientTape() as tape:
    logits = model(x_batch, training=True)
    loss_value = loss_fn(y_batch, logits)
    grads = tape.gradient(loss_value, model.trainable_variables)
    with open(os.path.join(gradient_location, f"grads_{job_index}.pickle"), "wb") as file:
        pickle.dump(grads, file)

