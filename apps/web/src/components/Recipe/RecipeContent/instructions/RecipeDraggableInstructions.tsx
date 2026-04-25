'use client';
import React from 'react';
import { models_Instruction } from '@4ks/api-fetch';
import { DragDropContext, Draggable } from '@hello-pangea/dnd';
import { default as Droppable } from '@/components/StrictModeDroppable';
import { DropResult } from '@hello-pangea/dnd';
import RecipeInstruction from './RecipeInstruction';

type RecipeInstructionsProps = {
  data: models_Instruction[];
  onDragEnd: (result: DropResult) => void;
  onChange: (index: number, data: models_Instruction) => void;
  onDelete: (index: number) => void;
};

export default function RecipeDraggableInstructions({
  data,
  onDragEnd,
  onChange,
  onDelete,
}: RecipeInstructionsProps) {
  return (
    <DragDropContext onDragEnd={onDragEnd}>
      <Droppable droppableId="instructions">
        {(provided) => (
          <ul
            style={{ listStyleType: 'none', paddingInlineStart: '0px' }}
            className="instructions"
            {...provided.droppableProps}
            ref={provided.innerRef}
          >
            {data?.map((instruction, index) => {
              const key = `instruction_${index}_${instruction.id}`;
              return (
                <Draggable key={key} draggableId={key} index={index}>
                  {(provided, snapshot) => (
                    <li
                      ref={provided.innerRef}
                      {...provided.draggableProps}
                      {...provided.dragHandleProps}
                    >
                      <RecipeInstruction
                        isDragging={snapshot.isDragging}
                        index={index}
                        key={instruction.text}
                        data={instruction}
                        handleInstructionDelete={onDelete}
                        handleInstructionChange={onChange}
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
