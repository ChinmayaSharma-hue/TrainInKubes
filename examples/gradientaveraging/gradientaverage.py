import tensorflow as tf
import os
import pickle

model_location = os.environ['MODEL_LOCATION']
gradient_location = os.environ['GRADIENT_LOCATION']
numberOfGrads = os.environ['NUMBER_OF_GRADS']

grads_list = []

# Picle load all the grad files from the gradient location and add them to a list
for i in range(int(numberOfGrads)):
    with open(gradient_location + '/grads_' + str(i) + '.pickle', 'rb') as f:
        grads_list.append(pickle.load(f))

# Find the average of all the gradients
avg_grads = [tf.reduce_mean([g[i] for g in grads_list], axis=0) for i in range(len(grads_list[0]))]

# Define an optimizer that can update the model
optimizer = tf.keras.optimizers.SGD(learning_rate=0.01)

# Load the model
model = tf.keras.models.load_model(model_location)

# Apply the gradients to the model
optimizer.apply_gradients(zip(avg_grads, model.trainable_variables))

# Save the model
model.save(model_location)