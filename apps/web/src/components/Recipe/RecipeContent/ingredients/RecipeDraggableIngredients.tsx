'use client';
import React from 'react';
import { models_Ingredient } from '@4ks/api-fetch';
import { DragDropContext, Draggable } from '@hello-pangea/dnd';
import { default as Droppable } from '@/components/StrictModeDroppable';
import { DropResult } from '@hello-pangea/dnd';

const RecipeIngredient = React.lazy(() => import('./RecipeIngredient'));

type RecipeIngredientsProps = {
  data: models_Ingredient[];
  onDragEnd: (result: DropResult) => void;
  onChange: (index: number, data: models_Ingredient) => void;
  onDelete: (index: number) => void;
};

export default function RecipeDraggableIngredients({
  data,
  onDragEnd,
  onChange,
  onDelete,
}: RecipeIngredientsProps) {
  return (
    <DragDropContext onDragEnd={onDragEnd}>
      <Droppable droppableId="ingredients">
        {(provided) => (
          <ul
            style={{ listStyleType: 'none', paddingInlineStart: '0px' }}
            className="ingredients"
            {...provided.droppableProps}
            ref={provided.innerRef}
          >
            {data?.map((ingredient, index) => {
              const key = `ingredient_${index}_${ingredient.id}`;
              return (
                <Draggable key={key} draggableId={key} index={index}>
                  {(provided, snapshot) => (
                    <li
                      key={ingredient.name}
                      ref={provided.innerRef}
                      {...provided.draggableProps}
                      {...provided.dragHandleProps}
                    >
                      <RecipeIngredient
                        index={index}
                        data={ingredient}
                        isDragging={snapshot.isDragging}
                        handleIngredientDelete={onDelete}
                        handleIngredientChange={onChange}
                      />
                    </li>
                  )}
                </Draggable>
              );
            })}
            {provided.placeholder}
          </ul>
        )}
      </Droppable>
    </DragDropContext>
  );
}
